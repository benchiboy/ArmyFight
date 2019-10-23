package afplaydetail

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

const (
	SQL_NEWDB   = "NewDB  ===>"
	SQL_INSERT  = "Insert ===>"
	SQL_UPDATE  = "Update ===>"
	SQL_SELECT  = "Select ===>"
	SQL_DELETE  = "Delete ===>"
	SQL_ELAPSED = "Elapsed===>"
	SQL_ERROR   = "Error  ===>"
	SQL_TITLE   = "===================================="
	DEBUG       = 1
	INFO        = 2
)

type Search struct {
	AutoId     int64  `json:"auto_id"`
	BatchNo    string `json:"batch_no"`
	PlayNo     int64  `json:"play_no"`
	Player     string `json:"player"`
	PlayCard   string `json:"play_card"`
	CoinCnt    int64  `json:"coin_cnt"`
	Result     string `json:"result"`
	InsertDate int64  `json:"insert_date"`
	UpdateDate int64  `json:"update_date"`
	PageNo     int    `json:"page_no"`
	PageSize   int    `json:"page_size"`
	ExtraWhere string `json:"extra_where"`
	SortFld    string `json:"sort_fld"`
}

type AfPlayDetailList struct {
	DB            *sql.DB
	Level         int
	Total         int            `json:"total"`
	AfPlayDetails []AfPlayDetail `json:"AfPlayDetail"`
}

type AfPlayDetail struct {
	AutoId     int64  `json:"auto_id"`
	BatchNo    string `json:"batch_no"`
	PlayNo     int64  `json:"play_no"`
	Player     string `json:"player"`
	PlayCard   string `json:"play_card"`
	CoinCnt    int64  `json:"coin_cnt"`
	Result     string `json:"result"`
	InsertDate int64  `json:"insert_date"`
	UpdateDate int64  `json:"update_date"`
}

type Form struct {
	Form AfPlayDetail `json:"AfPlayDetail"`
}

/*
	说明：创建实例对象
	入参：db:数据库sql.DB, 数据库已经连接, level:日志级别
	出参：实例对象
*/

func New(db *sql.DB, level int) *AfPlayDetailList {
	if db == nil {
		log.Println(SQL_SELECT, "Database is nil")
		return nil
	}
	return &AfPlayDetailList{DB: db, Total: 0, AfPlayDetails: make([]AfPlayDetail, 0), Level: level}
}

/*
	说明：创建实例对象
	入参：url:连接数据的url, 数据库还没有CONNECTED, level:日志级别
	出参：实例对象
*/

func NewUrl(url string, level int) *AfPlayDetailList {
	var err error
	db, err := sql.Open("mysql", url)
	if err != nil {
		log.Println(SQL_SELECT, "Open database error:", err)
		return nil
	}
	if err = db.Ping(); err != nil {
		log.Println(SQL_SELECT, "Ping database error:", err)
		return nil
	}
	return &AfPlayDetailList{DB: db, Total: 0, AfPlayDetails: make([]AfPlayDetail, 0), Level: level}
}

/*
	说明：得到符合条件的总条数
	入参：s: 查询条件
	出参：参数1：返回符合条件的总条件, 参数2：如果错误返回错误对象
*/

func (r *AfPlayDetailList) GetTotal(s Search) (int, error) {
	var where string
	l := time.Now()

	if s.AutoId != 0 {
		where += " and auto_id=" + fmt.Sprintf("%d", s.AutoId)
	}

	if s.BatchNo != "" {
		where += " and batch_no='" + s.BatchNo + "'"
	}

	if s.PlayNo != 0 {
		where += " and play_no=" + fmt.Sprintf("%d", s.PlayNo)
	}

	if s.Player != "" {
		where += " and player='" + s.Player + "'"
	}

	if s.PlayCard != "" {
		where += " and play_card='" + s.PlayCard + "'"
	}

	if s.CoinCnt != 0 {
		where += " and coin_cnt=" + fmt.Sprintf("%d", s.CoinCnt)
	}

	if s.Result != "" {
		where += " and result='" + s.Result + "'"
	}

	if s.InsertDate != 0 {
		where += " and insert_date=" + fmt.Sprintf("%d", s.InsertDate)
	}

	if s.UpdateDate != 0 {
		where += " and update_date=" + fmt.Sprintf("%d", s.UpdateDate)
	}

	if s.ExtraWhere != "" {
		where += s.ExtraWhere
	}

	qrySql := fmt.Sprintf("Select count(1) as total from af_play_detail   where 1=1 %s", where)
	if r.Level == DEBUG {
		log.Println(SQL_SELECT, qrySql)
	}
	rows, err := r.DB.Query(qrySql)
	if err != nil {
		log.Println(SQL_ERROR, err.Error())
		return 0, err
	}
	defer rows.Close()
	var total int
	for rows.Next() {
		rows.Scan(&total)
	}
	if r.Level == DEBUG {
		log.Println(SQL_ELAPSED, time.Since(l))
	}
	return total, nil
}

/*
	说明：根据主键查询符合条件的条数
	入参：s: 查询条件
	出参：参数1：返回符合条件的对象, 参数2：如果错误返回错误对象
*/

func (r AfPlayDetailList) Get(s Search) (*AfPlayDetail, error) {
	var where string
	l := time.Now()

	if s.AutoId != 0 {
		where += " and auto_id=" + fmt.Sprintf("%d", s.AutoId)
	}

	if s.BatchNo != "" {
		where += " and batch_no='" + s.BatchNo + "'"
	}

	if s.PlayNo != 0 {
		where += " and play_no=" + fmt.Sprintf("%d", s.PlayNo)
	}

	if s.Player != "" {
		where += " and player='" + s.Player + "'"
	}

	if s.PlayCard != "" {
		where += " and play_card='" + s.PlayCard + "'"
	}

	if s.CoinCnt != 0 {
		where += " and coin_cnt=" + fmt.Sprintf("%d", s.CoinCnt)
	}

	if s.Result != "" {
		where += " and result='" + s.Result + "'"
	}

	if s.InsertDate != 0 {
		where += " and insert_date=" + fmt.Sprintf("%d", s.InsertDate)
	}

	if s.UpdateDate != 0 {
		where += " and update_date=" + fmt.Sprintf("%d", s.UpdateDate)
	}

	if s.ExtraWhere != "" {
		where += s.ExtraWhere
	}

	qrySql := fmt.Sprintf("Select auto_id,batch_no,play_no,player,play_card,coin_cnt,result,insert_date,update_date, from af_play_detail where 1=1 %s ", where)
	if r.Level == DEBUG {
		log.Println(SQL_SELECT, qrySql)
	}
	rows, err := r.DB.Query(qrySql)
	if err != nil {
		log.Println(SQL_ERROR, err.Error())
		return nil, err
	}
	defer rows.Close()

	var p AfPlayDetail
	if !rows.Next() {
		return nil, fmt.Errorf("Not Finded Record")
	} else {
		err := rows.Scan(&p.AutoId, &p.BatchNo, &p.PlayNo, &p.Player, &p.PlayCard, &p.CoinCnt, &p.Result, &p.InsertDate, &p.UpdateDate)
		if err != nil {
			log.Println(SQL_ERROR, err.Error())
			return nil, err
		}
	}
	log.Println(SQL_ELAPSED, r)
	if r.Level == DEBUG {
		log.Println(SQL_ELAPSED, time.Since(l))
	}
	return &p, nil
}

/*
	说明：根据条件查询复核条件对象列表，支持分页查询
	入参：s: 查询条件
	出参：参数1：返回符合条件的对象列表, 参数2：如果错误返回错误对象
*/

func (r *AfPlayDetailList) GetList(s Search) ([]AfPlayDetail, error) {
	var where string
	l := time.Now()

	if s.AutoId != 0 {
		where += " and auto_id=" + fmt.Sprintf("%d", s.AutoId)
	}

	if s.BatchNo != "" {
		where += " and batch_no='" + s.BatchNo + "'"
	}

	if s.PlayNo != 0 {
		where += " and play_no=" + fmt.Sprintf("%d", s.PlayNo)
	}

	if s.Player != "" {
		where += " and player='" + s.Player + "'"
	}

	if s.PlayCard != "" {
		where += " and play_card='" + s.PlayCard + "'"
	}

	if s.CoinCnt != 0 {
		where += " and coin_cnt=" + fmt.Sprintf("%d", s.CoinCnt)
	}

	if s.Result != "" {
		where += " and result='" + s.Result + "'"
	}

	if s.InsertDate != 0 {
		where += " and insert_date=" + fmt.Sprintf("%d", s.InsertDate)
	}

	if s.UpdateDate != 0 {
		where += " and update_date=" + fmt.Sprintf("%d", s.UpdateDate)
	}

	if s.ExtraWhere != "" {
		where += s.ExtraWhere
	}

	var qrySql string
	if s.PageSize == 0 && s.PageNo == 0 {
		qrySql = fmt.Sprintf("Select auto_id,batch_no,play_no,player,play_card,coin_cnt,result,insert_date,update_date, from af_play_detail where 1=1 %s", where)
	} else {
		qrySql = fmt.Sprintf("Select auto_id,batch_no,play_no,player,play_card,coin_cnt,result,insert_date,update_date, from af_play_detail where 1=1 %s Limit %d offset %d", where, s.PageSize, (s.PageNo-1)*s.PageSize)
	}
	if r.Level == DEBUG {
		log.Println(SQL_SELECT, qrySql)
	}
	rows, err := r.DB.Query(qrySql)
	if err != nil {
		log.Println(SQL_ERROR, err.Error())
		return nil, err
	}
	defer rows.Close()

	var p AfPlayDetail
	for rows.Next() {
		rows.Scan(&p.AutoId, &p.BatchNo, &p.PlayNo, &p.Player, &p.PlayCard, &p.CoinCnt, &p.Result, &p.InsertDate, &p.UpdateDate)
		r.AfPlayDetails = append(r.AfPlayDetails, p)
	}
	log.Println(SQL_ELAPSED, r)
	if r.Level == DEBUG {
		log.Println(SQL_ELAPSED, time.Since(l))
	}
	return r.AfPlayDetails, nil
}

/*
	说明：根据主键查询符合条件的记录，并保持成MAP
	入参：s: 查询条件
	出参：参数1：返回符合条件的对象, 参数2：如果错误返回错误对象
*/

func (r *AfPlayDetailList) GetExt(s Search) (map[string]string, error) {
	var where string
	l := time.Now()

	if s.AutoId != 0 {
		where += " and auto_id=" + fmt.Sprintf("%d", s.AutoId)
	}

	if s.BatchNo != "" {
		where += " and batch_no='" + s.BatchNo + "'"
	}

	if s.PlayNo != 0 {
		where += " and play_no=" + fmt.Sprintf("%d", s.PlayNo)
	}

	if s.Player != "" {
		where += " and player='" + s.Player + "'"
	}

	if s.PlayCard != "" {
		where += " and play_card='" + s.PlayCard + "'"
	}

	if s.CoinCnt != 0 {
		where += " and coin_cnt=" + fmt.Sprintf("%d", s.CoinCnt)
	}

	if s.Result != "" {
		where += " and result='" + s.Result + "'"
	}

	if s.InsertDate != 0 {
		where += " and insert_date=" + fmt.Sprintf("%d", s.InsertDate)
	}

	if s.UpdateDate != 0 {
		where += " and update_date=" + fmt.Sprintf("%d", s.UpdateDate)
	}

	qrySql := fmt.Sprintf("Select auto_id,batch_no,play_no,player,play_card,coin_cnt,result,insert_date,update_date, from af_play_detail where 1=1 %s ", where)
	if r.Level == DEBUG {
		log.Println(SQL_SELECT, qrySql)
	}
	rows, err := r.DB.Query(qrySql)
	if err != nil {
		log.Println(SQL_ERROR, err.Error())
		return nil, err
	}
	defer rows.Close()

	Columns, _ := rows.Columns()

	values := make([]sql.RawBytes, len(Columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	if !rows.Next() {
		return nil, fmt.Errorf("Not Finded Record")
	} else {
		err = rows.Scan(scanArgs...)
	}

	fldValMap := make(map[string]string)
	for k, v := range Columns {
		fldValMap[v] = string(values[k])
	}

	log.Println(SQL_ELAPSED, "==========>>>>>>>>>>>", fldValMap)
	if r.Level == DEBUG {
		log.Println(SQL_ELAPSED, time.Since(l))
	}
	return fldValMap, nil

}

/*
	说明：插入对象到数据表中，这个方法要求对象的各个属性必须赋值
	入参：p:插入的对象
	出参：参数1：如果出错，返回错误对象；成功返回nil
*/

func (r AfPlayDetailList) Insert(p AfPlayDetail) error {
	l := time.Now()
	exeSql := fmt.Sprintf("Insert into  af_play_detail(auto_id,batch_no,play_no,player,play_card,coin_cnt,result,insert_date,update_date,)  values(?,?,?,?,?,?,?,?,?,)")
	if r.Level == DEBUG {
		log.Println(SQL_INSERT, exeSql)
	}
	_, err := r.DB.Exec(exeSql, p.AutoId, p.BatchNo, p.PlayNo, p.Player, p.PlayCard, p.CoinCnt, p.Result, p.InsertDate, p.UpdateDate)
	if err != nil {
		log.Println(SQL_ERROR, err.Error())
		return err
	}
	if r.Level == DEBUG {
		log.Println(SQL_ELAPSED, time.Since(l))
	}
	return nil
}

/*
	说明：插入对象到数据表中，这个方法会判读对象的各个属性，如果属性不为空，才加入插入列中；
	入参：p:插入的对象
	出参：参数1：如果出错，返回错误对象；成功返回nil
*/

func (r AfPlayDetailList) InsertEntity(p AfPlayDetail, tr *sql.Tx) error {
	l := time.Now()
	var colNames, colTags string
	valSlice := make([]interface{}, 0)

	if p.AutoId != 0 {
		colNames += "auto_id,"
		colTags += "?,"
		valSlice = append(valSlice, p.AutoId)
	}

	if p.BatchNo != "" {
		colNames += "batch_no,"
		colTags += "?,"
		valSlice = append(valSlice, p.BatchNo)
	}

	if p.PlayNo != 0 {
		colNames += "play_no,"
		colTags += "?,"
		valSlice = append(valSlice, p.PlayNo)
	}

	if p.Player != "" {
		colNames += "player,"
		colTags += "?,"
		valSlice = append(valSlice, p.Player)
	}

	if p.PlayCard != "" {
		colNames += "play_card,"
		colTags += "?,"
		valSlice = append(valSlice, p.PlayCard)
	}

	if p.CoinCnt != 0 {
		colNames += "coin_cnt,"
		colTags += "?,"
		valSlice = append(valSlice, p.CoinCnt)
	}

	if p.Result != "" {
		colNames += "result,"
		colTags += "?,"
		valSlice = append(valSlice, p.Result)
	}

	if p.InsertDate != 0 {
		colNames += "insert_date,"
		colTags += "?,"
		valSlice = append(valSlice, p.InsertDate)
	}

	if p.UpdateDate != 0 {
		colNames += "update_date,"
		colTags += "?,"
		valSlice = append(valSlice, p.UpdateDate)
	}

	colNames = strings.TrimRight(colNames, ",")
	colTags = strings.TrimRight(colTags, ",")
	exeSql := fmt.Sprintf("Insert into  af_play_detail(%s)  values(%s)", colNames, colTags)
	if r.Level == DEBUG {
		log.Println(SQL_INSERT, exeSql)
	}

	var stmt *sql.Stmt
	var err error
	if tr == nil {
		stmt, err = r.DB.Prepare(exeSql)
	} else {
		stmt, err = tr.Prepare(exeSql)
	}
	if err != nil {
		log.Println(SQL_ERROR, err.Error())
		return err
	}
	defer stmt.Close()

	ret, err := stmt.Exec(valSlice...)
	if err != nil {
		log.Println(SQL_INSERT, "Insert data error: %v\n", err)
		return err
	}
	if LastInsertId, err := ret.LastInsertId(); nil == err {
		log.Println(SQL_INSERT, "LastInsertId:", LastInsertId)
	}
	if RowsAffected, err := ret.RowsAffected(); nil == err {
		log.Println(SQL_INSERT, "RowsAffected:", RowsAffected)
	}

	if r.Level == DEBUG {
		log.Println(SQL_ELAPSED, time.Since(l))
	}
	return nil
}

/*
	说明：插入一个MAP到数据表中；
	入参：m:插入的Map
	出参：参数1：如果出错，返回错误对象；成功返回nil
*/

func (r AfPlayDetailList) InsertMap(m map[string]interface{}, tr *sql.Tx) error {
	l := time.Now()
	var colNames, colTags string
	valSlice := make([]interface{}, 0)
	for k, v := range m {
		colNames += k + ","
		colTags += "?,"
		valSlice = append(valSlice, v)
	}
	colNames = strings.TrimRight(colNames, ",")
	colTags = strings.TrimRight(colTags, ",")

	exeSql := fmt.Sprintf("Insert into  af_play_detail(%s)  values(%s)", colNames, colTags)
	if r.Level == DEBUG {
		log.Println(SQL_INSERT, exeSql)
	}

	var stmt *sql.Stmt
	var err error
	if tr == nil {
		stmt, err = r.DB.Prepare(exeSql)
	} else {
		stmt, err = tr.Prepare(exeSql)
	}

	if err != nil {
		log.Println(SQL_ERROR, err.Error())
		return err
	}
	defer stmt.Close()

	ret, err := stmt.Exec(valSlice...)
	if err != nil {
		log.Println(SQL_INSERT, "insert data error: %v\n", err)
		return err
	}
	if LastInsertId, err := ret.LastInsertId(); nil == err {
		log.Println(SQL_INSERT, "LastInsertId:", LastInsertId)
	}
	if RowsAffected, err := ret.RowsAffected(); nil == err {
		log.Println(SQL_INSERT, "RowsAffected:", RowsAffected)
	}

	if r.Level == DEBUG {
		log.Println(SQL_ELAPSED, time.Since(l))
	}
	return nil
}

/*
	说明：插入对象到数据表中，这个方法会判读对象的各个属性，如果属性不为空，才加入插入列中；
	入参：p:插入的对象
	出参：参数1：如果出错，返回错误对象；成功返回nil
*/

func (r AfPlayDetailList) UpdataEntity(keyNo string, p AfPlayDetail, tr *sql.Tx) error {
	l := time.Now()
	var colNames string
	valSlice := make([]interface{}, 0)

	if p.AutoId != 0 {
		colNames += "auto_id=?,"
		valSlice = append(valSlice, p.AutoId)
	}

	if p.BatchNo != "" {
		colNames += "batch_no=?,"

		valSlice = append(valSlice, p.BatchNo)
	}

	if p.PlayNo != 0 {
		colNames += "play_no=?,"
		valSlice = append(valSlice, p.PlayNo)
	}

	if p.Player != "" {
		colNames += "player=?,"

		valSlice = append(valSlice, p.Player)
	}

	if p.PlayCard != "" {
		colNames += "play_card=?,"

		valSlice = append(valSlice, p.PlayCard)
	}

	if p.CoinCnt != 0 {
		colNames += "coin_cnt=?,"
		valSlice = append(valSlice, p.CoinCnt)
	}

	if p.Result != "" {
		colNames += "result=?,"

		valSlice = append(valSlice, p.Result)
	}

	if p.InsertDate != 0 {
		colNames += "insert_date=?,"
		valSlice = append(valSlice, p.InsertDate)
	}

	if p.UpdateDate != 0 {
		colNames += "update_date=?,"
		valSlice = append(valSlice, p.UpdateDate)
	}

	colNames = strings.TrimRight(colNames, ",")
	valSlice = append(valSlice, keyNo)

	exeSql := fmt.Sprintf("update  af_play_detail  set %s  where auto_id=? ", colNames)
	if r.Level == DEBUG {
		log.Println(SQL_INSERT, exeSql)
	}

	var stmt *sql.Stmt
	var err error
	if tr == nil {
		stmt, err = r.DB.Prepare(exeSql)
	} else {
		stmt, err = tr.Prepare(exeSql)
	}

	if err != nil {
		log.Println(SQL_ERROR, err.Error())
		return err
	}
	defer stmt.Close()

	ret, err := stmt.Exec(valSlice...)
	if err != nil {
		log.Println(SQL_INSERT, "Update data error: %v\n", err)
		return err
	}
	if LastInsertId, err := ret.LastInsertId(); nil == err {
		log.Println(SQL_INSERT, "LastInsertId:", LastInsertId)
	}
	if RowsAffected, err := ret.RowsAffected(); nil == err {
		log.Println(SQL_INSERT, "RowsAffected:", RowsAffected)
	}

	if r.Level == DEBUG {
		log.Println(SQL_ELAPSED, time.Since(l))
	}
	return nil
}

/*
	说明：根据更新主键及更新Map值更新数据表；
	入参：keyNo:更新数据的关键条件，m:更新数据列的Map
	出参：参数1：如果出错，返回错误对象；成功返回nil
*/

func (r AfPlayDetailList) UpdateMap(batchNo string, playNo int64, player string, m map[string]interface{}, tr *sql.Tx) error {
	l := time.Now()

	var colNames string
	valSlice := make([]interface{}, 0)
	for k, v := range m {
		colNames += k + "=?,"
		valSlice = append(valSlice, v)
	}
	valSlice = append(valSlice, batchNo)
	valSlice = append(valSlice, playNo)
	valSlice = append(valSlice, player)

	colNames = strings.TrimRight(colNames, ",")
	updateSql := fmt.Sprintf("Update af_play_detail set %s where batch_no=? and play_no=? and player=?", colNames)
	if r.Level == DEBUG {
		log.Println(SQL_UPDATE, updateSql)
	}
	var stmt *sql.Stmt
	var err error
	if tr == nil {
		stmt, err = r.DB.Prepare(updateSql)
	} else {
		stmt, err = tr.Prepare(updateSql)
	}

	if err != nil {
		log.Println(SQL_ERROR, err.Error())
		return err
	}
	fmt.Println(valSlice)
	ret, err := stmt.Exec(valSlice...)
	if err != nil {
		log.Println(SQL_UPDATE, "Update data error: %v\n", err)
		return err
	}
	defer stmt.Close()

	if LastInsertId, err := ret.LastInsertId(); nil == err {
		log.Println(SQL_UPDATE, "LastInsertId:", LastInsertId)
	}
	if RowsAffected, err := ret.RowsAffected(); nil == err {
		log.Println(SQL_UPDATE, "RowsAffected:", RowsAffected)
	}
	if r.Level == DEBUG {
		log.Println(SQL_ELAPSED, time.Since(l))
	}
	return nil
}

/*
	说明：根据主键删除一条数据；
	入参：keyNo:要删除的主键值
	出参：参数1：如果出错，返回错误对象；成功返回nil
*/

func (r AfPlayDetailList) Delete(keyNo string, tr *sql.Tx) error {
	l := time.Now()
	delSql := fmt.Sprintf("Delete from  af_play_detail  where auto_id=?")
	if r.Level == DEBUG {
		log.Println(SQL_UPDATE, delSql)
	}

	var stmt *sql.Stmt
	var err error
	if tr == nil {
		stmt, err = r.DB.Prepare(delSql)
	} else {
		stmt, err = tr.Prepare(delSql)
	}

	if err != nil {
		log.Println(SQL_ERROR, err.Error())
		return err
	}
	ret, err := stmt.Exec(keyNo)
	if err != nil {
		log.Println(SQL_DELETE, "Delete error: %v\n", err)
		return err
	}
	defer stmt.Close()

	if LastInsertId, err := ret.LastInsertId(); nil == err {
		log.Println(SQL_DELETE, "LastInsertId:", LastInsertId)
	}
	if RowsAffected, err := ret.RowsAffected(); nil == err {
		log.Println(SQL_DELETE, "RowsAffected:", RowsAffected)
	}
	if r.Level == DEBUG {
		log.Println(SQL_ELAPSED, time.Since(l))
	}
	return nil
}
