package utils

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/samber/lo"

	"github.com/go-playground/validator/v10"
)

var ValidationMessages = map[string]func(e validator.FieldError) string{
	"required": func(e validator.FieldError) string { return "This field is required." },
	"email":    func(e validator.FieldError) string { return "Invalid email format." },
	"min":      func(e validator.FieldError) string { return "Value must be at least " + e.Param() + "." },
	"max":      func(e validator.FieldError) string { return "Value must be at most " + e.Param() + "." },
	"password": func(e validator.FieldError) string {
		return "Password must contain upper, lower, number, and special character."
	},
	// Add more as needed
}

func isDate(fl validator.FieldLevel) bool {
	date, err := time.Parse("2006-01-02", fl.Field().String())
	return err == nil && !date.IsZero()
}

func iso8601Time(fl validator.FieldLevel) bool {
	_, err := time.Parse(time.RFC3339, fl.Field().String())
	return err == nil
}

func passwordValidator(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := true

	// Iterate through each rune in the password and apply rule checks
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
			// case unicode.IsPunct(char) || unicode.IsSymbol(char):
			// 	hasSpecial = true
		}
	}

	// Ensure all conditions are satisfied
	return len(password) >= 8 && hasUpper && hasLower && hasNumber && hasSpecial
}

func imageValidator(fl validator.FieldLevel) bool {
	image := fl.Field().String()
	match, _ := regexp.MatchString(".*\\.(jpg|jpeg|png|gif|svg|webp)$", image)
	return match
}

func websiteValidator(fl validator.FieldLevel) bool {
	website := fl.Field().String()
	match, _ := regexp.MatchString("^(http|https)://.*", website)
	return match
}

func postcodeValidator(fl validator.FieldLevel) bool {
	postcode := fl.Field().String()
	match, _ := regexp.MatchString("^[0-9]{3}-[0-9]{4}$", postcode)
	return match
}

func isVerifyCode(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	match, _ := regexp.MatchString("^[0-9]{6}$", value)
	return match
}

func isUniqueByKeys(fl validator.FieldLevel) bool {
	field := fl.Field()

	if field.Kind() != reflect.Slice {
		return false
	}

	keyNames := strings.Fields(fl.Param())
	found := make(map[string]bool)

	for i := 0; i < field.Len(); i++ {
		item := field.Index(i)
		if item.Kind() != reflect.Struct {
			continue
		}

		var keyValues []string
		for _, keyName := range keyNames {
			key := item.FieldByName(keyName)
			if !key.IsValid() {
				return false
			}
			keyValues = append(keyValues, fmt.Sprintf("%v", key.Interface()))
		}

		compositeKey := strings.Join(keyValues, "|")
		if found[compositeKey] {
			return false
		}
		found[compositeKey] = true
	}

	return true
}

func isBase64Image(fl validator.FieldLevel) bool {
	base64String := fl.Field().String()

	allowedTypes := []string{}
	if len(fl.Param()) > 0 {
		allowedTypes = strings.Split(fl.Param(), " ")
	}

	prefixes := []string{
		"data:image/png;base64,",
		"data:image/jpeg;base64,",
		"data:image/gif;base64,",
		"data:image/webp;base64,",
		"data:image/svg+xml;base64,",
	}
	for _, prefix := range prefixes {
		base64String = strings.TrimPrefix(base64String, prefix)
	}

	// decode base64 string
	data, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return false
	}

	contentType := ""
	if strings.HasPrefix(string(data), "<svg") {
		contentType = "image/svg+xml"
	} else {
		contentType = http.DetectContentType(data)
	}

	if IsEmpty(allowedTypes) {
		return strings.HasPrefix(contentType, "image/")
	}

	return slices.Contains(allowedTypes, contentType)
}

func minfileSizeMatch(fl validator.FieldLevel) bool {
	content := fl.Field().String()
	minSize, err := strconv.Atoi(fl.Param())
	if err != nil {
		return false
	}

	totalSize := []byte(content)

	return len(totalSize) >= minSize
}

func maxfileSizeMatch(fl validator.FieldLevel) bool {
	content := fl.Field().String()
	maxSize, err := strconv.Atoi(fl.Param())
	if err != nil {
		return false
	}

	totalSize := []byte(content)

	return len(totalSize) <= maxSize
}

func requiredNumber(fl validator.FieldLevel) bool {
	return fl.Field().CanInt()
}

func googleAdSnippetAsset(fl validator.FieldLevel) bool {
	values := fl.Field().Interface().([]string)
	for _, value := range values {
		if len(value) < 1 || len(value) > 25 {
			return false
		}
	}

	return true
}

func googleAdAssetLinkUrl(fl validator.FieldLevel) bool {
	u, err := url.ParseRequestURI(fl.Field().String())
	return err == nil && u.Scheme != "" && u.Host != ""
}

func googleAdFinalUrls(fl validator.FieldLevel) bool {
	values := fl.Field().Interface().([]string)
	for _, value := range values {
		u, err := url.ParseRequestURI(value)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return false
		}
	}

	return true
}

func googleAdKeyword(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	words := strings.Fields(value)

	return len(words) <= 10
}

func dateAfter(fl validator.FieldLevel) bool {
	field := fl.Param()
	if field == "" {
		return false
	}

	value := time.Now().Format(time.DateOnly)

	if fl.Parent().FieldByName(field).IsValid() {
		value = fl.Parent().FieldByName(field).String()
	}

	dateCompare, _ := time.Parse(time.DateOnly, value)
	datetime, _ := StringToTime(fl.Field().String(), time.DateOnly)

	return !datetime.Before(dateCompare)
}

func notInSlice(fl validator.FieldLevel) bool {
	fieldNameCheck := fl.Param() // Ids
	if fieldNameCheck == "" {
		return false
	}

	parent := fl.Parent()
	// check is slice
	fieldCheck := parent.FieldByName(fieldNameCheck)
	if !fieldCheck.IsValid() || fieldCheck.Kind() != reflect.Slice {
		return false
	}

	value := fl.Field()
	// validate type
	if fieldCheck.Len() > 0 && fieldCheck.Index(0).Type() != value.Type() {
		return false
	}

	// validate value
	for i := 0; i < fieldCheck.Len(); i++ {
		if reflect.DeepEqual(value.Interface(), fieldCheck.Index(i).Interface()) {
			return false
		}
	}
	return true
}

func validateEnums(fl validator.FieldLevel) bool {
	fieldValue := fl.Field()
	fieldType := fieldValue.Kind()
	if fieldType != reflect.Slice {
		return false
	}

	validValues := fl.Param()
	if validValues == "" {
		return false
	}
	validValuesSlice := strings.Split(validValues, " ")

	for i := 0; i < fieldValue.Len(); i++ {
		elem := fieldValue.Index(i).Interface()

		if !lo.Contains(validValuesSlice, fmt.Sprintf("%v", elem)) {
			return false
		}
	}

	return true
}

// Regex:https://regex101.com/r/TTvTO7/1
// Docs: https://apispec.billing-robo.jp/public/request_marunage_credit/bulk_register.html
func billingCodeValidator(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	pattern := `[A-Za-z0-9!"#$%&'()*+,\-./:;<=>?@[\\\]^_` + "`" + `{|}~]`
	return regexp.MustCompile(pattern).MatchString(value)
}

// Example: required_when=field1 value1 & field2 value2
// Memo: Only support `and` condition. Please, use `&` to separate conditions. In case of `or` condition, please, use `required_if` instead.
func requiredWhen(fl validator.FieldLevel) bool {
	conditions := strings.Split(fl.Param(), "&")
	parent := fl.Parent()

	isRequired := true
	for _, condition := range conditions {
		condition = strings.TrimSpace(condition)
		conditionParts := strings.Split(condition, " ")
		if len(conditionParts) != 2 {
			return false
		}

		fieldName := conditionParts[0]
		fieldValue := conditionParts[1]

		field := parent.FieldByName(fieldName)
		if !field.IsValid() {
			return false
		}

		if field.Kind() == reflect.String {
			isRequired = isRequired && field.String() == fieldValue
		} else if field.Kind() == reflect.Int {
			isRequired = isRequired && fmt.Sprintf("%d", field.Int()) == fieldValue
		} else if field.Kind() == reflect.Bool {
			isRequired = isRequired && fmt.Sprintf("%t", field.Bool()) == fieldValue
		} else {
			return false
		}
	}

	if !isRequired {
		return true
	}

	return isRequired && !lo.IsEmpty(fl.Field().String())
}

func slugValidator(fl validator.FieldLevel) bool {
	regex := regexp.MustCompile(`^[a-zA-Z0-9\-]+$`)
	return regex.MatchString(fl.Field().String())
}

func ValidateStruct(data interface{}) error {
	v := validator.New(validator.WithRequiredStructEnabled())
	_ = v.RegisterValidation("date", isDate)
	_ = v.RegisterValidation("iso8601Time", iso8601Time)
	_ = v.RegisterValidation("image", imageValidator)
	_ = v.RegisterValidation("password", passwordValidator)
	_ = v.RegisterValidation("postcode", postcodeValidator)
	_ = v.RegisterValidation("verifycode", isVerifyCode)
	_ = v.RegisterValidation("website", websiteValidator)
	_ = v.RegisterValidation("uniqueKeys", isUniqueByKeys)
	_ = v.RegisterValidation("base64image", isBase64Image)
	_ = v.RegisterValidation("minsize", minfileSizeMatch)
	_ = v.RegisterValidation("maxsize", maxfileSizeMatch)
	_ = v.RegisterValidation("required_number", requiredNumber)
	_ = v.RegisterValidation("google_ad_snippet_asset", googleAdSnippetAsset)
	_ = v.RegisterValidation("google_ad_asset_link_url", googleAdAssetLinkUrl)
	_ = v.RegisterValidation("google_ad_final_urls", googleAdFinalUrls)
	_ = v.RegisterValidation("google_ad_key_word", googleAdKeyword)
	_ = v.RegisterValidation("date_after", dateAfter)
	_ = v.RegisterValidation("not_in_slice", notInSlice)
	_ = v.RegisterValidation("enums", validateEnums)
	_ = v.RegisterValidation("billing_code", billingCodeValidator)
	_ = v.RegisterValidation("required_when", requiredWhen)
	_ = v.RegisterValidation("slug", slugValidator)

	return v.Struct(data)
}

func ValidateSingle(val any, rules string) error {
	v := validator.New(validator.WithRequiredStructEnabled())

	return v.Var(val, rules)
}
