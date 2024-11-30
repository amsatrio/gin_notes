package util

import (
	"fmt"
	"reflect"
	"regexp"

	"gorm.io/gorm"

	"github.com/amsatrio/gin_notes/model/request"
)

func ApplyPaginate(pageInt int, limitInt int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := 0
		if pageInt >= 0 {
			offset = (pageInt) * limitInt
		}
		return db.Offset(offset).Limit(limitInt)
	}
}

func ApplySorting(db *gorm.DB, sorts []request.Sort) *gorm.DB {
	for _, sort := range sorts {
		if sort.Desc {
			return db.Order(sort.Id + " DESC")
		}
		return db.Order(sort.Id + " ASC")
	}
	return db
}

func ApplyFiltering(db *gorm.DB, filter []request.Filter) *gorm.DB {
	for _, condition := range filter {
		key, value, operation := condition.Id, condition.Value, condition.MatchMode

		valueString := ""

		if str, ok := value.(string); ok {
			fmt.Println("key: " + key + "; value: " + str + "; operation: " + operation.String())
			valueString = str
		}

		switch operation {
		case request.EQUALS:
			db = db.Where(key+" = ?", value)
		default:
			db = db.Where(key+" LIKE ?", "%"+valueString+"%")
		}
	}
	return db
}

func ApplyGlobalSearch(db *gorm.DB, search string, modelMap map[string]string) *gorm.DB {
	if search == "" {
		Log("INFO", "util", "ApplyGlobalSearch", "search is empty")
		return db
	}

	regex := regexp.MustCompile(`[^a-zA-Z0-9\s]+`)

	// cleansing search for security
	search = regex.ReplaceAllString(search, "")

	searchQuery := ""
	for key, value := range modelMap {
		// Log("INFO", "util", "ApplyGlobalSearch", "key: "+key+", value: "+value)
		if value != "string" {
			continue
		}

		key = CamelCaseToSnakeCase(key)

		if searchQuery != "" {
			searchQuery = searchQuery + " or "
		}
		searchQuery = searchQuery + key + " like " + "'%" + search + "%'"
	}
	if search != "" {
		db = db.Where(searchQuery)
	}
	return db
}

func GetJSONFieldTypes(s interface{}) map[string]string {
	fieldTypes := make(map[string]string)
	val := reflect.ValueOf(s)
	typ := reflect.TypeOf(s)

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("json")
		fieldType := val.Field(i).Type().String()

		fieldTypes[tag] = fieldType
	}

	return fieldTypes
}
