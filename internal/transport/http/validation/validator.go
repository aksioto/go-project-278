package validation

import (
	"net/url"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Init registers custom validators and configures tag name function for JSON field names.
func Init() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return fld.Name
		}
		return name
	})

	_ = RegisterCustomValidators(v) //nolint:errcheck // error is not critical for app startup
}

func RegisterCustomValidators(v *validator.Validate) error {
	return v.RegisterValidation("rfc3986url", validateRFC3986URL)
}

func validateRFC3986URL(fl validator.FieldLevel) bool {
	raw := fl.Field().String()

	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return false
	}

	// Ensure required parts are present
	if u.Scheme == "" || u.Host == "" {
		return false
	}

	// Allow only supported schemes
	switch u.Scheme {
	case "http", "https":
		return true
	default:
		return false
	}
}
