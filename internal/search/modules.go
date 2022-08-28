package search

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gitlab.com/edea-dev/edea-server/internal/model"
	"go.uber.org/zap"
)

type SearchParam struct {
	Field  string   `json:"field"`
	Op     string   `json:"op,omitempty"`
	Values []string `json:"values"`
}

type ParamValues struct {
	Field  string        `json:"field"`
	Values []interface{} `json:"values"`
}

func SearchModule(c *gin.Context) {
	var sp []SearchParam
	if err := c.BindJSON(&sp); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var sb strings.Builder
	var qc []interface{}

	// put the user condition first, just in case
	currentUser, _ := c.Keys["user"].(*model.User)
	if currentUser == nil {
		sb.WriteString("private = false")
	} else {
		sb.WriteString("(private = false OR user_id = ?)")
		qc = append(qc, currentUser.ID.String())
	}

	for i := 0; i < len(sp); i++ {
		paramsAdded := false
		isNumeric := false

		lenValues := len(sp[i].Values)

		// skip empty fields
		if lenValues == 0 {
			continue
		}

		// order of query parameters is field first then values
		qc = append(qc, sp[i].Field)

		// parse and add numeric parameters
		if num, err := strconv.ParseFloat(sp[i].Values[0], 64); err == nil {
			if lenValues == 1 {
				qc = append(qc, num)
				paramsAdded = true
				isNumeric = true
			} else {
				// try to parse all the values as numbers
				nums := make([]float64, lenValues)
				j := 0
				for ; j < lenValues; j++ {
					nums[j], err = strconv.ParseFloat(sp[i].Values[j], 64)
					if err != nil {
						break
					}
				}

				// only say we added the parameters if really all parsed successfully
				if j == lenValues {
					paramsAdded = true
					isNumeric = true
					qc = append(qc, nums)
				}
			}
		}

		// add string params
		if !paramsAdded {
			if lenValues == 1 {
				qc = append(qc, sp[i].Values[0])
			} else {
				qc = append(qc, sp[i].Values)
			}
		}

		sb.WriteString(" AND ")

		cast := "numeric"
		extract := "->"
		if !isNumeric {
			cast = "text"
			extract = "->>"
		}

		// append the query string for the parameter we're currently processing
		if lenValues > 1 {
			sb.WriteString(fmt.Sprintf("(metadata -> 'params' %s ?)::%s IN ?", extract, cast))
		} else {
			switch sp[i].Op {
			case "=":
				sb.WriteString(fmt.Sprintf("(metadata -> 'params' %s ?)::%s = ?", extract, cast))
			case "<":
				sb.WriteString(fmt.Sprintf("(metadata -> 'params' %s ?)::%s < ?", extract, cast))
			case ">":
				sb.WriteString(fmt.Sprintf("(metadata -> 'params' %s ?)::%s > ?", extract, cast))
			default:
				c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("invalid or missing op for param %s", sp[i].Field)})
				return
			}
		}

	}

	zap.S().Infof("query: %s, params: %#v", sb.String(), qc)

	// query the db with the selectors
	var modules []model.Module
	tx := model.DB.Model(&model.Module{}).Where(sb.String(), qc...).Preload("Category").Preload("User").Find(&modules)
	if err := tx.Error; err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, modules)
}

func Filters(c *gin.Context) {
	var filters []model.Filter
	tx := model.DB.Find(&filters)
	if tx.Error != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, tx.Error)
		return
	}

	c.JSON(http.StatusOK, filters)
}

func GetParametersForCategory(c *gin.Context) {
	var pgQuery = `
SELECT json_object_agg(field, values)
FROM (SELECT key AS field, array_agg(value) AS values
		FROM (SELECT f.key, f.value
			FROM (SELECT metadata -> 'params' AS params
					FROM modules
					WHERE metadata -> 'params'::text != 'null') t,
					jsonb_each(t.params) f
			WHERE t.params is not null
			GROUP BY f.value, f.key
			ORDER BY f.value) s
		GROUP BY key) j;
`

	var res string
	tx := model.DB.Raw(pgQuery).Scan(&res)
	if tx.Error != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, tx.Error)
		return
	}

	c.Data(http.StatusOK, "application/json", []byte(res))
}
