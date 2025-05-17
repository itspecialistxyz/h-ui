package controller

import (
	"fmt"
	"h-ui/model/constant"
	"h-ui/model/vo"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	_ = validate.RegisterValidation("validateStr", validateStr)
}

func validateStr(f validator.FieldLevel) bool {
	field := f.Field().String()
	// 字符串必须6-32位是字母或者数字或部分特殊字符的组合
	reg := "^[a-zA-Z0-9!@#$%^&*()_+-=]{6,32}$"
	compile := regexp.MustCompile(reg)
	return field == "" || compile.MatchString(field)
}

func validateField[T interface{}](c *gin.Context, field T) (T, error) {
	var bindErr error
	if c.Request.Method == http.MethodGet {
		bindErr = c.ShouldBindQuery(&field)
	} else if c.Request.Method == http.MethodPost ||
		c.Request.Method == http.MethodPut ||
		c.Request.Method == http.MethodDelete {
		bindErr = c.ShouldBindJSON(&field)
	}

	if bindErr != nil {
		vo.Fail(fmt.Sprintf("Invalid request format: %v", bindErr), c)
		return field, fmt.Errorf("binding error: %w", bindErr)
	}

	if err := validate.Struct(&field); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			vo.Fail(fmt.Sprintf("Validation failed: %v", validationErrors), c)
			return field, fmt.Errorf("validation error: %w", err)
		}
		vo.Fail(constant.InvalidError, c)
		return field, fmt.Errorf(constant.InvalidError)
	}
	return field, nil
}
