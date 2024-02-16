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
	id, err := getAliasFromPayload(payload)
	if err != nil {
		return []byte("Error getting path param"), err
	}
	budgetID, ok := f.budgetPathRelation[id]
	if !ok {
		return []byte(`{"error":"Budget not found"}`), errors.New("Budget not found")
	}

	log.Trace().Msg("Getting responce of budget")
	respBudget, err := f.getBudgetCurrentLimit(budgetID)
	if err != nil {
		return []byte(`{"error":"Error occured while getting budget"}`), errors.New("Error occured while getting budget")
	}
	log.Trace().Msg("Got responce of budget")

	respJson, err := anyToJson(respBudget)
	if err != nil {
		return []byte(`{"error":"Error occured while unmarshaling budget"}`), errors.New("Error occured while unmarshaling budget")
	}
	log.Trace().Msg(string(respJson))
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

/*
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
	ffib := fireflyiiiLimit{}
	if err := json.Unmarshal(body, &ffib); err != nil {
		return nil, err
	}
	return fireflyiiiBudgetToreturn(ffib), nil
}
*/

func (f FireflyiiiConnection) getBudgetCurrentLimit(id int) (*returnStruct, error) {
	path := fmt.Sprintf("/api/v1/budgets/%d/limits&start=%s", id, getFirstMonthDate())
	r, err := f.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	log.Trace().Msgf("Getting budget request %s", r.URL.String())
	resp, err := f.cl.Do(r)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	log.Trace().Msgf("Got responce body %s", string(body))

	ffib := fireflyiiiLimit{}
	if err := json.Unmarshal(body, &ffib); err != nil {
		return nil, err
	}
	return fireflyiiiBudgetToreturn(ffib), nil
}

func fireflyiiiBudgetToreturn(f fireflyiiiLimit) *returnStruct {
	spent, err := strconv.ParseFloat(f.Data[0].Attributes.Spent, 64)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert spent")
	}

	limit, err := strconv.ParseFloat(f.Data[0].Attributes.Amount, 64)
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert limit aka budgeted")
	}

	return &returnStruct{
		Type:        f.Data[0].Attributes.Period,
		Budgeted:    fmt.Sprintf("%.2f", math.Abs(limit)),
		Spent:       fmt.Sprintf("%.2f", math.Abs(spent)),
		LeftToSpent: fmt.Sprintf("%.2f", math.Abs(limit)-math.Abs(spent)),
	}
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

type fireflyiiiLimit struct {
	Data []struct {
		Attributes struct {
			Start  time.Time `json:"start"`
			Amount string    `json:"amount"`
			Period string    `json:"period"`
			Spent  string    `json:"spent"`
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
