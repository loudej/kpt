package fnruntime

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	goerrors "errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptrace"
	"net/textproto"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/GoogleContainerTools/kpt/internal/printer"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/gcrane"
	"github.com/google/go-containerregistry/pkg/name"
	cranev1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	specsv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"google.golang.org/api/option"
	run "google.golang.org/api/run/v1"
)

type cloudRunContainerFn struct {
	*ContainerFn
}

func init() {
    rand.Seed(time.Now().UnixNano())
}

const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
func randLetters(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}


func (c* cloudRunContainerFn) Run(reader io.Reader, writer io.Writer) error {
	ctx := c.ContainerFn.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	project := "rickon-fishfood"
	region := "us-central1"
	registry := fmt.Sprintf("us-docker.pkg.dev/%s/kpt-cloud-run", project)

	runService, err := run.NewService(ctx, option.WithEndpoint(fmt.Sprintf("https://%s-run.googleapis.com", region)))
	if err != nil {
		return err
	}

	startFind := time.Now()

	var service *run.Service

	serviceList, err := runService.Namespaces.Services.List(fmt.Sprintf("namespaces/%s", project)).Do()
	if err != nil {
		return err
	}

	for _, item := range serviceList.Items {	
		if item.Metadata != nil {
			if item.Metadata.Annotations != nil {
				if item.Metadata.Annotations["function-image"] == c.Image {
					service = item
					break
				}
			}
		}	
	}

	durationFind := time.Since(startFind)

	startCreate := time.Now()

	if service == nil {
		name := fmt.Sprintf("kpt-function-%s", randLetters(6))

		// RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
		// gcloud artifacts repositories create --repository-format=docker --project=rickon-fishfood --location=us kpt-cloud-run
		newImage := path.Join(registry, c.Image)
		createdImage, err := pushCloudRunImage(c.Image, newImage)
		if err!=nil {
			return err
		}

		newService := &run.Service{
			ApiVersion:      "serving.knative.dev/v1",
			Kind:            "Service",
			Metadata:        &run.ObjectMeta{
				Name: name,
				Annotations: map[string]string{
					"function-image": c.Image,
				},
			},
			Spec:            &run.ServiceSpec{
				Template:        &run.RevisionTemplate{
					Metadata:        &run.ObjectMeta{
						Name: fmt.Sprintf("%s-v1", name),
					},
					Spec:            &run.RevisionSpec{
						ContainerConcurrency: 0,
						Containers:           []*run.Container{
							{
								// Args:                     []string{},
								// Command:                  []string{},
								// Env:                      []*run.EnvVar{},
								// EnvFrom:                  []*run.EnvFromSource{},
								Image:                    createdImage,
								// ImagePullPolicy:          image,
								// LivenessProbe:            &run.Probe{},
								// Name:                     "",
								// Ports:                    []*run.ContainerPort{
								// 	{
								// 		ContainerPort:   8080,
								// 		Name:            "http",
								// 		Protocol:        "http",
								// 		// ForceSendFields: []string{},
								// 		// NullFields:      []string{},
								// 	},
								// },
								// ReadinessProbe:           &run.Probe{},
								// Resources:                &run.ResourceRequirements{},
								// SecurityContext:          &run.SecurityContext{},
								// StartupProbe:             &run.Probe{},
								// TerminationMessagePath:   "",
								// TerminationMessagePolicy: "",
								// VolumeMounts:             []*run.VolumeMount{},
								// WorkingDir:               "",
								// ForceSendFields:          []string{},
								// NullFields:               []string{},
							},
						},
						// ServiceAccountName:   "",
						// TimeoutSeconds:       0,
						// Volumes:              []*run.Volume{},
						// ForceSendFields:      []string{},
						// NullFields:           []string{},
					},
					// ForceSendFields: []string{},
					// NullFields:      []string{},
				},
				// Traffic:         []*run.TrafficTarget{},
				// ForceSendFields: []string{},
				// NullFields:      []string{},
			},
			// Status:          &run.ServiceStatus{},
			// ServerResponse:  googleapi.ServerResponse{},
			// ForceSendFields: []string{},
			// NullFields:      []string{},
		}

		createdService, err := runService.Namespaces.Services.Create(fmt.Sprintf("namespaces/%s", project), newService).Do()
		if err != nil {
			return fmt.Errorf("creating cloud run service: %w", err)
		}

		_, err = runService.Projects.Locations.Services.SetIamPolicy(
			fmt.Sprintf("projects/%s/locations/%s/services/%s", project, region, createdService.Metadata.Name),
			&run.SetIamPolicyRequest{
				Policy: &run.Policy{Bindings: []*run.Binding{{
					Members: []string{"allUsers"},
					Role:    "roles/run.invoker",
				}}},
			},
		).Do()
		if err != nil {
			return fmt.Errorf("set service access: %w", err)
		}

		if err := waitForReady(ctx, runService, region, project, createdService.Metadata.Name, "Ready"); err!=nil {
			return fmt.Errorf("wait for service ready: %v", err)
		}

		if err := waitForReady(ctx, runService, region, project, createdService.Metadata.Name, "RoutesReady"); err!=nil {
			return fmt.Errorf("wait for routes ready: %v", err)
		}
		
		if service, err = getService(runService, region, project, createdService.Metadata.Name); err!=nil {
			return err
		}
	}
	durationCreate := time.Since(startCreate)

	if service == nil {
		return fmt.Errorf("unable to process cloud run function")
	}

	startCall := time.Now()

	trace := &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			fmt.Printf("%v: GetConn: %+v\n", time.Since(startCall), hostPort)
		},
		GotConn: func(connInfo httptrace.GotConnInfo) {
			fmt.Printf("%v: GotConn: %+v\n", time.Since(startCall), connInfo)
		},
		PutIdleConn: func(err error) {
			fmt.Printf("%v: PutIdleConn: %+v\n", time.Since(startCall), err)
		},
		GotFirstResponseByte: func() {
			fmt.Printf("%v: GotFirstResponseByte\n", time.Since(startCall))
		},
		Got100Continue: func() {
			fmt.Printf("%v: Got100Continue\n", time.Since(startCall))
		},
		Got1xxResponse: func(code int, header textproto.MIMEHeader) error {
			fmt.Printf("%v: Got1xxResponse: %+v %+v\n", time.Since(startCall), code, header)
			return nil
		},
		DNSStart: func(x httptrace.DNSStartInfo) {
			fmt.Printf("%v: DNSStart: %+v\n", time.Since(startCall), x)
		},
		DNSDone: func(x httptrace.DNSDoneInfo) {
			fmt.Printf("%v: DNSDone: %+v\n", time.Since(startCall), x)
		},
		ConnectStart: func(network string, addr string) {
			fmt.Printf("%v: ConnectStart: %+v %+v\n", time.Since(startCall), network, addr)
		},
		ConnectDone: func(network string, addr string, err error) {
			fmt.Printf("%v: ConnectDone: %+v %+v %+v\n", time.Since(startCall), network, addr, err)
		},
		TLSHandshakeStart: func() {
			fmt.Printf("%v: TLSHandshakeStart\n", time.Since(startCall))
		},
		TLSHandshakeDone: func(x tls.ConnectionState, err error) {
			fmt.Printf("%v: TLSHandshakeDone: %+v %+v\n", time.Since(startCall), x, err)
		},
		WroteHeaderField: func(key string, value []string) {
			fmt.Printf("%v: WroteHeaderField: %+v %+v\n", time.Since(startCall), key, value)
		},
		WroteHeaders: func() {
			fmt.Printf("%v: WroteHeaders\n", time.Since(startCall))
		},
		Wait100Continue: func() {
			fmt.Printf("%v: Wait100Continue", time.Since(startCall))
		},
		WroteRequest: func(x httptrace.WroteRequestInfo) {
			fmt.Printf("%v: WroteRequest: %+v\n", time.Since(startCall), x)
		},
	}

	req, err := http.NewRequest("POST", service.Status.Url, reader)
	if err != nil {
		return fmt.Errorf("sending request %s: %w", service.Status.Url, err)
	}
	resp, err := http.DefaultTransport.RoundTrip(req.WithContext(httptrace.WithClientTrace(ctx, trace)))
	if err != nil {
		return fmt.Errorf("sending request %s: %w", service.Status.Url, err)
	}


	// resp, err := http.Post(service.Status.Url, "text/vnd.yaml", reader)

	// TODO(dejardin) error for non-200
	
	if _,err:=io.Copy(writer,resp.Body); err!=nil {
		return fmt.Errorf("receiving response %s: %w", service.Status.Url, err)
	}

	durationCall := time.Since(startCall)

	fmt.Fprintf(os.Stderr, "time spent %s find:%s create:%s call:%s\n", service.Status.Url, durationFind, durationCreate, durationCall)

	return nil	
}

func pushCloudRunImage(baseRef string, newTag string) (string,error) {
	annotate := true

	// Find and open executable
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}

	executableFile, err := os.Open(ex)
	if err != nil {
		return "", err
	}
	defer executableFile.Close()


	// Get base image
	options := &[]crane.Option{crane.WithAuthFromKeychain(gcrane.Keychain)}

	base, err := crane.Pull(baseRef, *options...)
	if err != nil {
		return "", fmt.Errorf("pulling %s: %v", baseRef, err)
	}

	// Make new layer
	tarFile, err := ioutil.TempFile("", "tar")
	if err != nil {
		return "", fmt.Errorf("temp tar file: %v", err)
	}
	defer os.Remove(tarFile.Name())

	if err := func() error {
		defer tarFile.Close()

		gw := gzip.NewWriter(tarFile)
		defer gw.Close()

		tw := tar.NewWriter(gw)
		defer tw.Close()

		if err := addFile(tw, executableFile, "kpt", 0755); err != nil {
			return err
		}

		return nil
	}(); err != nil {
		return "", err
	}

	// Append new layer
	newLayers := []string{tarFile.Name()}
	img, err := crane.Append(base, newLayers...)
	if err != nil {
		return "", fmt.Errorf("appending %v: %v", newLayers, err)
	}
	
	if annotate {
		ref, err := name.ParseReference(baseRef)
		if err != nil {
			return "", fmt.Errorf("parsing ref %q: %v", baseRef, err)
		}

		baseDigest, err := base.Digest()
		if err != nil {
			return "", err
		}
		anns := map[string]string{
			specsv1.AnnotationAuthors: baseDigest.String(),
		}
		if _, ok := ref.(name.Tag); ok {
			anns[specsv1.AnnotationBaseImageName] = ref.Name()
		}
		img = mutate.Annotations(img, anns).(cranev1.Image)
	}

	cfg, err := img.ConfigFile()
	if err != nil {
		return "", fmt.Errorf("getting config: %v", err)
	}
	cfg = cfg.DeepCopy()

	// Set entrypoint.
	cfg.Config.Entrypoint = append([]string{"/lib/ld-musl-x86_64.so.1", "/kpt", "serve"}, cfg.Config.Entrypoint...)

	// Mutate image.
	img, err = mutate.Config(img, cfg.Config)
	if err != nil {
		return "", fmt.Errorf("mutating config: %v", err)
	}


	if err := crane.Push(img, newTag, *options...); err != nil {
		return "", fmt.Errorf("pushing image %s: %v", newTag, err)
	}

	ref, err := name.ParseReference(newTag)
	if err != nil {
		return "", fmt.Errorf("parsing reference %s: %v", newTag, err)
	}

	d, err := img.Digest()
	if err != nil {
		return "", fmt.Errorf("digest: %v", err)
	}

	return ref.Context().Digest(d.String()).String(), nil
}

func addFile(tw*tar.Writer, file*os.File, name string, mode int64) error {

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(stat, "")
	if err != nil {
		return err
	}

	header.Name = name
	header.Mode = mode
	header.Uid = 0
	header.Gid = 0
	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}

	return nil
}

func ListenAndServeFunction(args []string) error {
	
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		// TODO(dejardin) see getDockerCmd for addtl setop, e.g. passing env vars
		ctx := req.Context()

		if err := func() error {
			cmd := exec.CommandContext(ctx, args[0], args[1:]...)

			errSink := bytes.Buffer{}
			cmd.Stdin = req.Body
			cmd.Stdout = res
			cmd.Stderr = &errSink
			
			if err := cmd.Run(); err != nil {
				var exitErr *exec.ExitError
				if goerrors.As(err, &exitErr) {
					return &ExecError{
						OriginalErr:    exitErr,
						ExitCode:       exitErr.ExitCode(),
						Stderr:         errSink.String(),
						TruncateOutput: printer.TruncateOutput,
					}
				}
				return fmt.Errorf("unexpected function error: %w", err)
			}

			// TODO(dejardin) send as "trailers"?
			// if errSink.Len() > 0 {
			// 	f.FnResult.Stderr = errSink.String()
			// }

			return nil
		}(); err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
	})
	return http.ListenAndServe(":8080", nil)

}

func waitForReady(ctx context.Context, c *run.APIService, region, project, name, condition string) error {
	t := time.NewTicker(time.Second * 5)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			svc, err := getService(c, region, project, name)
			if err != nil {
				return fmt.Errorf("failed to query service for readiness: %w", err)
			}
			for _, c := range svc.Status.Conditions {
				if c.Type == condition {
					if c.Status == "True" {
						return nil
					} else if c.Status == "False" {
						return fmt.Errorf("service could not become %q (status:%s) (reason:%s) %s",
							condition, c.Status, c.Reason, c.Message)
					}
				}
			}
		}
	}
}

func getService(c *run.APIService, region, project, name string) (*run.Service, error) {
	return c.Namespaces.Services.Get(fmt.Sprintf("namespaces/%s/services/%s", project, name)).Do()
}
