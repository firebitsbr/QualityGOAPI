package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/mux"
)

type customer struct {
	Id             string    `json:"id"`
	CodCli         string    `json:"codcli"`
	Cedula         string    `json:"cedula"`
	Nombre_1       string    `json:"nombre_1"`
	Nombre_2       string    `json:"nombre_2"`
	Apellido_1     string    `json:"apellido_1"`
	Apellido_2     string    `json:"apellido_2"`
	Nombre_Com     string    `json:"nombre_com"`
	Nombre_Bus     string    `json:"nombre_bus"`
	Nombre_Cal     string    `json:"nombre_cal"`
	Telefono_1     string    `json:"telefono_1"`
	Telefono_2     string    `json:"telefono_2"`
	Celular_1      string    `json:"celular_1"`
	Celular_2      string    `json:"celular_2"`
	TELS           string    `json:"tels"`
	Direccion      string    `json:"direccion"`
	Regimen        string    `json:"regimen"`
	EMail          string    `json:"email"`
	RegistraFecNac bool      `json:"registra_fec_nac"`
	FecNac         time.Time `json:"fecnac"`
	CodMcpio       string    `json:"codmcpio"`
	CodDpto        string    `json:"coddpto"`
	TipCap         int8      `json:"tipcap"`
	TipID          string    `json:"tipid"`
	CodList        string    `json:"codlist"`
	FechaRegistro  time.Time `json:"fecha_registro"`
	MargenReteICA  float32   `json:"margen_rete_ica"`
	RetAnyBase     bool      `json:"ret_any_base"`
	CodVen         string    `json:"codven"`
	CodZona        string    `json:"codzona"`
	PlazoCR        int8      `json:"plazo_cr"`
	ExentoIVA      bool      `json:"exento_iva"`
	Activo         bool      `json:"activo"`
	MotivoBloqueo  string    `json:"motivo_bloqueo"`
	CodNeg         string    `json:"codneg"`
}

type pagination struct {
	Total     uint64 `json:"total"`
	Page_Size int    `json:"page_size"`
	Page_No   int    `json:"page_no"`
}

type customers struct {
	Pagination pagination `json:"pagination"`
	Customers  []customer `json:"data"`
}

var DATABASE_NAME string

func (Customer_v2) TableName() string {
	return DATABASE_NAME + ".dbo.ven_clientes"
}

func GetCustomers(c *appContext, w http.ResponseWriter, r *http.Request) (int, error) {

	var sQuery_TMPL, sQuery_Counter string
	var sQuery, sDBPrefix bytes.Buffer
	var sFilter_Query, sFilter_Pagination string = "", ""
	var cust customer

	custmrs := &customers{}
	custmrs.Customers = make([]customer, 0)

	// Set DataBaseName
	DATABASE_NAME = r.Header.Get("host_database")

	// Valid End-Points
	// /customers -> All Customers
	// /customers/90 -> Customer Id = 90
	// /customers?pagen_no=1&page_size=50 -> Customer's selection from 1 to 50
	// /customers?pagen_no=2&page_size=50 -> Customer's selection from 51 to 100
	// /customers?pagen_no=1&page_size=50&filter=CARLOS -> Customer's selection from 1 to 50 Where customer's name
	// contains "CARLOS" string.

	vars := mux.Vars(r)
	query := r.URL.Query()

	sId, sFilter, sPage_size, sOffset, sPageNo, host_database := vars["id"], query.Get("filter"), query.Get("page_size"),
		query.Get("offset"), query.Get("page_no"), r.Header.Get("host_database") // Check if Id is provided

	// Low Cost concatenation process
	sDBPrefix.WriteString(host_database)
	sDBPrefix.WriteString(".DBO.")

	// Query Values
	type query_values struct {
		Id         int
		Page_Size  int
		OffSet     int
		DBName     string
		Filter     string
		Pagination string
	}

	// Casting
	id, _ := strconv.Atoi(sId)
	page_size, _ := strconv.Atoi(sPage_size)
	offset, _ := strconv.Atoi(sOffset)
	page_no, _ := strconv.Atoi(sPageNo)

	// Set Filter
	if strings.Trim(sFilter, " ") != "" {
		sFilter_Query = "AND CLI.Nombre_Com LIKE '%" + strings.Replace(sFilter, " ", "%", -1) + "%'"
	}

	// OffSet, Page_Size and Page No.
	if strings.Trim(sPage_size, " ") != "" && strings.Trim(sPageNo, " ") != "" {
		sFilter_Pagination = "OFFSET " + strconv.Itoa((page_no-1)*page_size) + " ROWS \n" +
			"FETCH NEXT " + strconv.Itoa(page_size) + " ROWS ONLY"
	}

	// Fill Values
	substitute := query_values{Id: id, Page_Size: page_size, OffSet: offset, DBName: sDBPrefix.String(), Filter: sFilter_Query, Pagination: sFilter_Pagination}

	// Query selection
	switch {
	case strings.Trim(sId, " ") != "": // If a customer id is provided
		sQuery_TMPL = `
				SELECT Ven_Clientes.Id, Ven_Clientes.Cedula, Ven_Clientes.CodCli, CLI.Nombre_1, CLI.Nombre_2, CLI.Apellido_1, CLI.Apellido_2,
				CLI.NOMBRE_COM, CLI.NOMBRE_CAL, CLI.NOMBRE_BUS,
				CLI.Telefono_1, CLI.Telefono_2, CLI.Celular_1, CLI.Celular_2, CLI.TELS, CLI.Direccion, CLI.Regimen, CLI.EMail,
				CLI.RegistraFecNac, ISNULL(CLI.FecNac,GetDate()) As FecNac, CLI.CodMcpio, CLI.CodDpto, CLI.TipCap, CLI.TipID, LTRIM(RTRIM(Ven_Clientes.CODLIST)) As CodList,
				ISNULL(CLI.FechaRegistro,GetDate()) As FechaRegistro, Ven_Clientes.MARGENRETEICA, Ven_Clientes.RETANYBASE, Ven_Clientes.CodVen, Ven_Clientes.CodZona,
				Ven_Clientes.PlazoCR, Ven_Clientes.ExentoIVA, Ven_Clientes.Activo, Ven_Clientes.MotivoBloqueo, Ven_Clientes.CodNeg
				FROM {{.DBName}}Ven_Clientes
				LEFT JOIN {{.DBName}}Cnt_Terceros CLI ON CLI.CodTer = Ven_Clientes.Cedula
				WHERE Ven_Clientes.Id={{.Id}}
				ORDER BY CLI.Nombre_Com`
	default:
		sQuery_TMPL = `
				SELECT Ven_Clientes.Id, Ven_Clientes.Cedula, Ven_Clientes.CodCli, CLI.Nombre_1, CLI.Nombre_2, CLI.Apellido_1, CLI.Apellido_2,
				CLI.NOMBRE_COM, CLI.NOMBRE_CAL, CLI.NOMBRE_BUS,
				CLI.Telefono_1, CLI.Telefono_2, CLI.Celular_1, CLI.Celular_2, CLI.TELS, CLI.Direccion, CLI.Regimen, CLI.EMail,
				CLI.RegistraFecNac, ISNULL(CLI.FecNac,GetDate()) As FecNac, CLI.CodMcpio, CLI.CodDpto, CLI.TipCap, CLI.TipID, LTRIM(RTRIM(Ven_Clientes.CODLIST)) As CodList,
				ISNULL(CLI.FechaRegistro,GetDate()) As FechaRegistro, Ven_Clientes.MARGENRETEICA, Ven_Clientes.RETANYBASE, Ven_Clientes.CodVen, Ven_Clientes.CodZona,
				Ven_Clientes.PlazoCR, Ven_Clientes.ExentoIVA, Ven_Clientes.Activo, Ven_Clientes.MotivoBloqueo, Ven_Clientes.CodNeg
				FROM {{.DBName}}Ven_Clientes
				LEFT JOIN {{.DBName}}Cnt_Terceros CLI ON CLI.CodTer = Ven_Clientes.Cedula
				WHERE LTRIM(RTRIM(CLI.CodTer))<>'' {{.Filter}}
				ORDER BY CLI.Nombre_Com
				{{.Pagination}}
				`
	}

	// Setup template
	tmpL, err := template.New("sQuery_TMPL").Parse(sQuery_TMPL)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	err = tmpL.Execute(&sQuery, substitute)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Obtengo la conexión a la base de datos
	db, ok := c.dbs[r.Header.Get("host_domain")]

	if ok {

		rows, err := db.Raw(sQuery.String()).Rows()

		if err != nil {
			return http.StatusInternalServerError, err
		}

		for rows.Next() {
			err := rows.Scan(&cust.Id, &cust.Cedula, &cust.CodCli, &cust.Nombre_1, &cust.Nombre_2, &cust.Apellido_1, &cust.Apellido_2,
				&cust.Nombre_Com, &cust.Nombre_Bus, &cust.Nombre_Cal, &cust.Telefono_1, &cust.Telefono_2,
				&cust.Celular_1, &cust.Celular_2, &cust.TELS, &cust.Direccion, &cust.Regimen, &cust.EMail,
				&cust.RegistraFecNac, &cust.FecNac, &cust.CodMcpio, &cust.CodDpto, &cust.TipCap, &cust.TipID,
				&cust.CodList, &cust.FechaRegistro, &cust.MargenReteICA, &cust.RetAnyBase, &cust.CodVen,
				&cust.CodZona, &cust.PlazoCR, &cust.ExentoIVA, &cust.Activo, &cust.MotivoBloqueo, &cust.CodNeg)

			if err != nil {
				return http.StatusInternalServerError, err
			}
			custmrs.Customers = append(custmrs.Customers, cust)
		}

		// Get pagination info
		sQuery_Counter = fmt.Sprintf("SELECT Count(Id) As Total FROM %vVen_Clientes CLI WHERE Cedula<>'' %v", host_database+".DBO.", sFilter_Query)

		rows, err = db.Raw(sQuery_Counter).Rows()
		if err != nil {
			return http.StatusInternalServerError, err
		}

		var counter uint64
		if rows.Next() {
			err = rows.Scan(&counter)
			if err != nil {
				return http.StatusInternalServerError, err
			}
		}

		custmrs.Pagination.Total = counter
		custmrs.Pagination.Page_Size = page_size
		custmrs.Pagination.Page_No = page_no
		json.NewEncoder(w).Encode(custmrs)
	}
	return http.StatusOK, nil
}

type Customer_v2 struct {
	Id            int32   `json:"id,string"gorm:"column:id"gorm:"primary_key"`
	CodCli        string  `json:"codcli"gorm:"column:codcli"`
	Cedula        string  `json:"cedula"`
	CodList       string  `json:"codlist"gorm:"column:codlist"`
	Margenreteica float32 `json:"margen_rete_ica,string"gorm:"column:margenreteica"`
	Retanybase    *bool   `json:"ret_any_base"`
	CodVen        string  `json:"codven"gorm:"column:codven"`
	CodZona       string  `json:"codzona"gorm:"column:codzona"`
	PlazoCR       int8    `json:"plazo_cr"gorm:"column:plazocr"`
	ExentoIVA     *bool   `json:"exento_iva"gorm:"column:exentoiva"`
	Activo        *bool   `json:"activo"`
	MotivoBloqueo string  `json:"motivo_bloqueo"gorm:"column:motivobloqueo"`
	CodNeg        string  `json:"codneg"gorm:"column:codneg"`
}

// Create New Customer
func PostCustomers(c *appContext, w http.ResponseWriter, r *http.Request) (int, error) {
	var client Customer_v2
	// Set DataBaseName
	DATABASE_NAME = r.Header.Get("host_database")
	// Obtengo el cuerpo del body
	err := json.NewDecoder(r.Body).Decode(&client)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	// Obtengo la conexión a la base de datos
	db, ok := c.dbs[r.Header.Get("host_domain")]

	if ok {
		if dbc := db.Create(&client) ; dbc.Error != nil {
			return http.StatusInternalServerError, dbc.Error
		}
	}
	return http.StatusOK, nil
}

// Update existing Customer
func PutCustomers (c *appContext, w http.ResponseWriter, r *http.Request) (int, error) {
	var client Customer_v2
	var params Customer_v2
	var id = mux.Vars(r)["id"]
	// Set DataBaseName
	DATABASE_NAME = r.Header.Get("host_database")
	// Obtengo el cuerpo del body
	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	// Obtengo la conexión a la base de datos
	db, ok := c.dbs[r.Header.Get("host_domain")]
	//if ok {
	//	db.First(&client, id)
	//	fmt.Println(params)
	//	if dbc := db.Model(&client).Updates(map[string]interface{}{"CodList":params.CodList, "Margenreteica":params.Margenreteica,
	//	"Retanybase":params.Retanybase, "CodVen":params.CodVen, "CodZona":params.CodZona, "PlazoCR":params.PlazoCR, "ExentoIVA":params.ExentoIVA,
	//	"Activo":params.Activo, "MotivoBloqueo":params.MotivoBloqueo, "CodNeg":params.CodNeg}); dbc.Error != nil {
	//		return http.StatusInternalServerError, dbc.Error
	//	}
	//}
	if ok {
		db.First(&client, id)
		//fmt.Println(params)
		if dbc := db.Model(&client).Updates(&params); dbc.Error != nil {
			return http.StatusInternalServerError, dbc.Error
		}
	}
	return http.StatusOK, nil
}