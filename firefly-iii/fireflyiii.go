package fireflyiii

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

type FireflyiiiConnection struct {
	cl                 *http.Client
	PATToken           string
	FireflyiiiURL      string
	budgetPathRelation map[string]int
}

func NewFireflyiiiConnection(PAT, FireflyiiiURL string, BudgetPathRelation map[string]int) *FireflyiiiConnection {
	return &FireflyiiiConnection{
		cl:                 &http.Client{Timeout: time.Second * 30},
		PATToken:           PAT,
		FireflyiiiURL:      FireflyiiiURL,
		budgetPathRelation: BudgetPathRelation,
	}
}

func (f FireflyiiiConnection) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	fmt.Println(string(payload))
	id, err := getAliasFromPayload(payload)
	if err != nil {
		return []byte("Error getting path param"), err
	}
	budgetID, ok := f.budgetPathRelation[id]
	if !ok {
		return []byte(`{"error":"Budget not found"}`), errors.New("Budget not found")
	}

	log.Trace().Msg("Getting responce of budget")
	respBudget, err := f.getRespBudget(budgetID)
	if err != nil {
		return []byte(`{"error":"Error occured while getting budget"}`), errors.New("Error occured while getting budget")
	}
	log.Trace().Msg("Got responce of budget")

	respJson, err := anyToJson(respBudget)
	if err != nil {
		return []byte(`{"error":"Error occured while unmarshaling budget"}`), errors.New("Error occured while unmarshaling budget")
	}

	return respJson, nil
}

func getAliasFromPayload(payload []byte) (string, error) {
	obj := struct {
		PathParameters struct {
			Id string `json:"id"`
		} `json:"pathParameters"`
	}{}
	if err := json.Unmarshal(payload, &obj); err != nil {
		return "", errors.New("Failed to unmarshal params")
	}
	return obj.PathParameters.Id, nil
}

func (f FireflyiiiConnection) getRespBudget(id int) (*returnStruct, error) {
	req, err := f.newRequest(http.MethodGet, "/api/v1/budgets/"+fmt.Sprint(id)+
		fmt.Sprintf("?start=%s&end=%s", getFirstMonthDate(), getToday()), nil)
	if err != nil {
		return nil, err
	}
	log.Trace().Msgf("Getting budget request %s", req.URL.String())

	r, err := f.cl.Do(req)
	if err != nil {
		return nil, err
	}
	if r.StatusCode != http.StatusOK {
		return nil, errors.New("Failed to fetch with status " + r.Status)
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	log.Trace().Msgf("Got responce body %s", string(body))
	ffib := fireflyiiiBudget{}
	if err := json.Unmarshal(body, &ffib); err != nil {
		return nil, err
	}
	return fireflyiiiBudgetToreturn(ffib), nil
}

func fireflyiiiBudgetToreturn(f fireflyiiiBudget) *returnStruct {
	spent, err := getSpent(f)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert spent")
	}
	budgeted, err := strconv.ParseFloat(f.Data.Attributes.AutoBudgetAmount, 64)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert budgeted")
	}
	return &returnStruct{
		Type:        f.Data.Attributes.AutoBudgetPeriod,
		Budgeted:    fmt.Sprintf("%.2f", math.Abs(budgeted)),
		Spent:       fmt.Sprintf("%.2f", math.Abs(spent)),
		LeftToSpent: fmt.Sprintf("%.2f", math.Abs(budgeted)-math.Abs(spent)),
	}
}

func getSpent(f fireflyiiiBudget) (float64, error) {
	if len(f.Data.Attributes.Spent) == 0 {
		return 0.0, nil
	}
	return strconv.ParseFloat(f.Data.Attributes.Spent[0].Sum, 64)
}

func (f *FireflyiiiConnection) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, f.FireflyiiiURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/vnd.api+json")
	req.Header.Add("Authorization", "Bearer "+f.PATToken)
	return req, nil
}

type fireflyiiiBudget struct {
	Data struct {
		Attributes struct {
			Name             string `json:"name"`
			AutoBudgetAmount string `json:"auto_budget_amount"`
			AutoBudgetPeriod string `json:"auto_budget_period"`
			Spent            []struct {
				Sum string `json:"sum"`
			} `json:"spent"`
		} `json:"attributes"`
	} `json:"data"`
}

type returnStruct struct {
	Type        string `json:"type"`
	Spent       string `json:"spent"`
	LeftToSpent string `json:"left"`
	Budgeted    string `json:"budgeted"`
}

func anyToJson(a interface{}) ([]byte, error) {
	jsonBody, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return jsonBody, nil
}

func getToday() string {
	y, m, d := time.Now().Date()
	return fmt.Sprintf("%d-%d-%d", y, int(m), d)
}

func getFirstMonthDate() string {
	y, m, _ := time.Now().Date()
	return fmt.Sprintf("%d-%d-%d", y, int(m), 1)

}
