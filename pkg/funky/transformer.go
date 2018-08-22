package funky

import (
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type RequestTransformer interface {
	Transform(req *http.Request) (*Request, error)
}

type DefaultRequestTransformer struct {
	timeout      int
	secrets      []string
	secretClient v1.SecretInterface
	rw           HTTPReaderWriter
}

// NewDefaultRequestTransformer transforms a request into a format suitable for Dispatch language servers
func NewDefaultRequestTransformer(timeout int, secrets []string, client v1.SecretInterface, rw HTTPReaderWriter) RequestTransformer {
	return &DefaultRequestTransformer{
		timeout:      timeout,
		secrets:      secrets,
		secretClient: client,
		rw:           rw,
	}
}

func (r *DefaultRequestTransformer) Transform(req *http.Request) (*Request, error) {
	var body Request

	// populate function timeout
	ctx := map[string]interface{}{
		"timeout": r.timeout,
	}

	// populate secrets
	if len(r.secrets) != 0 {
		secretList, err := r.secretClient.List(metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		secrets := map[string]string{}
		for _, secretName := range r.secrets {
			for _, secretItem := range secretList.Items {
				if secretItem.GetName() != secretName {
					continue
				}

				for key, value := range secretItem.Data {
					secrets[key] = string(value)
				}
			}
		}

		ctx["secrets"] = secrets
	}

	// populate request attributes
	ctx["request"] = map[string]interface{}{
		"uri":    req.RequestURI,
		"method": req.Method,
		"header": req.Header,
	}

	body.Context = ctx

	var err error
	// populate payload
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}

	switch contentType {
	case "application/base64":
		var payload []byte
		err = r.rw.Read(&payload, req)
		body.Payload = payload
	case "text/plain":
		var payload string
		err = r.rw.Read(&payload, req)
		body.Payload = payload
	case "application/json":
		fallthrough
	case "application/yaml":
		var payload map[string]interface{}
		err = r.rw.Read(&payload, req)
		body.Payload = payload
	default:
		err = UnsupportedMediaTypeError(contentType)
	}

	if err != nil {
		return nil, err
	}

	return &body, nil
}

type ResponseTransformer interface {
	Transform(res Message, w http.ResponseWriter)
}

type DefaultResponseTransformer struct {
	rw HTTPReaderWriter
}

func NewDefaultResponseTransfomer(rw HTTPReaderWriter) ResponseTransformer {
	return &DefaultResponseTransformer{
		rw: rw,
	}
}

func (r *DefaultResponseTransformer) Transform(res Message, w http.ResponseWriter) {

}

type SecretInterfaceFactory interface {
	CreateSecretInterface(namespace string) v1.SecretInterface
}

type DefaultSecretInterfaceFactory struct {
	k8sClient *kubernetes.Clientset
}

func NewDefaultSecretInterfaceFactory(config *rest.Config) (SecretInterfaceFactory, error) {
	if config == nil {
		if c, err := rest.InClusterConfig(); err != nil {
			config = c
		} else {
			return nil, err
		}
	}

	clientset := kubernetes.NewForConfigOrDie(config)

	return &DefaultSecretInterfaceFactory{
		k8sClient: clientset,
	}, nil
}

func (f DefaultSecretInterfaceFactory) CreateSecretInterface(namespace string) v1.SecretInterface {
	return f.k8sClient.CoreV1().Secrets(namespace)
}
