package afplay

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
	PlayName   string `json:"play_name"`
	FromPlayer string `json:"from_player"`
	ToPlayer   string `json:"to_player"`
	CoinCnt    int64  `json:"coin_cnt"`
	MedalCnt   int64  `json:"medal_cnt"`
	Status     string `json:"status"`
	EndType    string `json:"end_type"`
	Winner     string `json:"winner"`
	InsertDate int64  `json:"insert_date"`
	UpdateDate int64  `json:"update_date"`
	Version    int64  `json:"version"`
	PageNo     int    `json:"page_no"`
	PageSize   int    `json:"page_size"`
	ExtraWhere string `json:"extra_where"`
	SortFld    string `json:"sort_fld"`
}

type AfPlayList struct {
	DB      *sql.DB
	Level   int
	Total   int      `json:"total"`
	AfPlays []AfPlay `json:"AfPlay"`
}

type AfPlay struct {
	AutoId     int64  `json:"auto_id"`
	BatchNo    string `json:"batch_no"`
	PlayName   string `json:"play_name"`
	FromPlayer string `json:"from_player"`
	ToPlayer   string `json:"to_player"`
	CoinCnt    int64  `json:"coin_cnt"`
	MedalCnt   int64  `json:"medal_cnt"`
	Status     string `json:"status"`
	EndType    string `json:"end_type"`
	Winner     string `json:"winner"`
	InsertDate int64  `json:"insert_date"`
	UpdateDate int64  `json:"update_date"`
	Version    int64  `json:"version"`
}

type Form struct {
	Form AfPlay `json:"AfPlay"`
}

/*
	说明：创建实例对象
	入参：db:数据库sql.DB, 数据库已经连接, level:日志级别
	出参：实例对象
*/

func New(db *sql.DB, level int) *AfPlayList {
	if db == nil {
		log.Println(SQL_SELECT, "Database is nil")
		return nil
	}
	return &AfPlayList{DB: db, Total: 0, AfPlays: make([]AfPlay, 0), Level: level}
}

/*
	说明：创建实例对象
	入参：url:连接数据的url, 数据库还没有CONNECTED, level:日志级别
	出参：实例对象
*/

func NewUrl(url string, level int) *AfPlayList {
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
	return &AfPlayList{DB: db, Total: 0, AfPlays: make([]AfPlay, 0), Level: level}
}

/*
	说明：得到符合条件的总条数
	入参：s: 查询条件
	出参：参数1：返回符合条件的总条件, 参数2：如果错误返回错误对象
*/

func (r *AfPlayList) GetTotal(s Search) (int, error) {
	var where string
	l := time.Now()

	if s.AutoId != 0 {
		where += " and auto_id=" + fmt.Sprintf("%d", s.AutoId)
	}

	if s.BatchNo != "" {
		where += " and batch_no='" + s.BatchNo + "'"
	}

	if s.PlayName != "" {
		where += " and play_name='" + s.PlayName + "'"
	}

	if s.FromPlayer != "" {
		where += " and from_player='" + s.FromPlayer + "'"
	}

	if s.ToPlayer != "" {
		where += " and to_player='" + s.ToPlayer + "'"
	}

	if s.CoinCnt != 0 {
		where += " and coin_cnt=" + fmt.Sprintf("%d", s.CoinCnt)
	}

	if s.MedalCnt != 0 {
		where += " and medal_cnt=" + fmt.Sprintf("%d", s.MedalCnt)
	}

	if s.Status != "" {
		where += " and status='" + s.Status + "'"
	}

	if s.EndType != "" {
		where += " and end_type='" + s.EndType + "'"
	}

	if s.Winner != "" {
		where += " and winner='" + s.Winner + "'"
	}

	if s.InsertDate != 0 {
		where += " and insert_date=" + fmt.Sprintf("%d", s.InsertDate)
	}

	if s.UpdateDate != 0 {
		where += " and update_date=" + fmt.Sprintf("%d", s.UpdateDate)
	}

	if s.Version != 0 {
		where += " and version=" + fmt.Sprintf("%d", s.Version)
	}

	if s.ExtraWhere != "" {
		where += s.ExtraWhere
	}

	qrySql := fmt.Sprintf("Select count(1) as total from af_play   where 1=1 %s", where)
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

func (r AfPlayList) Get(s Search) (*AfPlay, error) {
	var where string
	l := time.Now()

	if s.AutoId != 0 {
		where += " and auto_id=" + fmt.Sprintf("%d", s.AutoId)
	}

	if s.BatchNo != "" {
		where += " and batch_no='" + s.BatchNo + "'"
	}

	if s.PlayName != "" {
		where += " and play_name='" + s.PlayName + "'"
	}

	if s.FromPlayer != "" {
		where += " and from_player='" + s.FromPlayer + "'"
	}

	if s.ToPlayer != "" {
		where += " and to_player='" + s.ToPlayer + "'"
	}

	if s.CoinCnt != 0 {
		where += " and coin_cnt=" + fmt.Sprintf("%d", s.CoinCnt)
	}

	if s.MedalCnt != 0 {
		where += " and medal_cnt=" + fmt.Sprintf("%d", s.MedalCnt)
	}

	if s.Status != "" {
		where += " and status='" + s.Status + "'"
	}

	if s.EndType != "" {
		where += " and end_type='" + s.EndType + "'"
	}

	if s.Winner != "" {
		where += " and winner='" + s.Winner + "'"
	}

	if s.InsertDate != 0 {
		where += " and insert_date=" + fmt.Sprintf("%d", s.InsertDate)
	}

	if s.UpdateDate != 0 {
		where += " and update_date=" + fmt.Sprintf("%d", s.UpdateDate)
	}

	if s.Version != 0 {
		where += " and version=" + fmt.Sprintf("%d", s.Version)
	}

	if s.ExtraWhere != "" {
		where += s.ExtraWhere
	}

	qrySql := fmt.Sprintf("Select auto_id,batch_no,play_name,from_player,to_player,coin_cnt,medal_cnt,status,end_type,winner,insert_date,update_date,version from af_play where 1=1 %s ", where)
	if r.Level == DEBUG {
		log.Println(SQL_SELECT, qrySql)
	}
	rows, err := r.DB.Query(qrySql)
	if err != nil {
		log.Println(SQL_ERROR, err.Error())
		return nil, err
	}
	defer rows.Close()

	var p AfPlay
	if !rows.Next() {
		return nil, fmt.Errorf("Not Finded Record")
	} else {
		err := rows.Scan(&p.AutoId, &p.BatchNo, &p.PlayName, &p.FromPlayer, &p.ToPlayer, &p.CoinCnt, &p.MedalCnt, &p.Status, &p.EndType, &p.Winner, &p.InsertDate, &p.UpdateDate, &p.Version)
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

func (r *AfPlayList) GetList(s Search) ([]AfPlay, error) {
	var where string
	l := time.Now()

	if s.AutoId != 0 {
		where += " and auto_id=" + fmt.Sprintf("%d", s.AutoId)
	}

	if s.BatchNo != "" {
		where += " and batch_no='" + s.BatchNo + "'"
	}

	if s.PlayName != "" {
		where += " and play_name='" + s.PlayName + "'"
	}

	if s.FromPlayer != "" {
		where += " and from_player='" + s.FromPlayer + "'"
	}

	if s.ToPlayer != "" {
		where += " and to_player='" + s.ToPlayer + "'"
	}

	if s.CoinCnt != 0 {
		where += " and coin_cnt=" + fmt.Sprintf("%d", s.CoinCnt)
	}

	if s.MedalCnt != 0 {
		where += " and medal_cnt=" + fmt.Sprintf("%d", s.MedalCnt)
	}

	if s.Status != "" {
		where += " and status='" + s.Status + "'"
	}

	if s.EndType != "" {
		where += " and end_type='" + s.EndType + "'"
	}

	if s.Winner != "" {
		where += " and winner='" + s.Winner + "'"
	}

	if s.InsertDate != 0 {
		where += " and insert_date=" + fmt.Sprintf("%d", s.InsertDate)
	}

	if s.UpdateDate != 0 {
		where += " and update_date=" + fmt.Sprintf("%d", s.UpdateDate)
	}

	if s.Version != 0 {
		where += " and version=" + fmt.Sprintf("%d", s.Version)
	}

	if s.ExtraWhere != "" {
		where += s.ExtraWhere
	}

	var qrySql string
	if s.PageSize == 0 && s.PageNo == 0 {
		qrySql = fmt.Sprintf("Select auto_id,batch_no,play_name,from_player,to_player,coin_cnt,medal_cnt,status,end_type,winner,insert_date,update_date,version from af_play where 1=1 %s", where)
	} else {
		qrySql = fmt.Sprintf("Select auto_id,batch_no,play_name,from_player,to_player,coin_cnt,medal_cnt,status,end_type,winner,insert_date,update_date,version from af_play where 1=1 %s Limit %d offset %d", where, s.PageSize, (s.PageNo-1)*s.PageSize)
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

	var p AfPlay
	for rows.Next() {
		rows.Scan(&p.AutoId, &p.BatchNo, &p.PlayName, &p.FromPlayer, &p.ToPlayer, &p.CoinCnt, &p.MedalCnt, &p.Status, &p.EndType, &p.Winner, &p.InsertDate, &p.UpdateDate, &p.Version)
		r.AfPlays = append(r.AfPlays, p)
	}
	log.Println(SQL_ELAPSED, r)
	if r.Level == DEBUG {
		log.Println(SQL_ELAPSED, time.Since(l))
	}
	return r.AfPlays, nil
}

/*
	说明：根据主键查询符合条件的记录，并保持成MAP
	入参：s: 查询条件
	出参：参数1：返回符合条件的对象, 参数2：如果错误返回错误对象
*/

func (r *AfPlayList) GetExt(s Search) (map[string]string, error) {
	var where string
	l := time.Now()

	if s.AutoId != 0 {
		where += " and auto_id=" + fmt.Sprintf("%d", s.AutoId)
	}

	if s.BatchNo != "" {
		where += " and batch_no='" + s.BatchNo + "'"
	}

	if s.PlayName != "" {
		where += " and play_name='" + s.PlayName + "'"
	}

	if s.FromPlayer != "" {
		where += " and from_player='" + s.FromPlayer + "'"
	}

	if s.ToPlayer != "" {
		where += " and to_player='" + s.ToPlayer + "'"
	}

	if s.CoinCnt != 0 {
		where += " and coin_cnt=" + fmt.Sprintf("%d", s.CoinCnt)
	}

	if s.MedalCnt != 0 {
		where += " and medal_cnt=" + fmt.Sprintf("%d", s.MedalCnt)
	}

	if s.Status != "" {
		where += " and status='" + s.Status + "'"
	}

	if s.EndType != "" {
		where += " and end_type='" + s.EndType + "'"
	}

	if s.Winner != "" {
		where += " and winner='" + s.Winner + "'"
	}

	if s.InsertDate != 0 {
		where += " and insert_date=" + fmt.Sprintf("%d", s.InsertDate)
	}

	if s.UpdateDate != 0 {
		where += " and update_date=" + fmt.Sprintf("%d", s.UpdateDate)
	}

	if s.Version != 0 {
		where += " and version=" + fmt.Sprintf("%d", s.Version)
	}

	qrySql := fmt.Sprintf("Select auto_id,batch_no,play_name,from_player,to_player,coin_cnt,medal_cnt,status,end_type,winner,insert_date,update_date,version from af_play where 1=1 %s ", where)
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

func (r AfPlayList) Insert(p AfPlay) error {
	l := time.Now()
	exeSql := fmt.Sprintf("Insert into  af_play(auto_id,batch_no,play_name,from_player,to_player,coin_cnt,medal_cnt,status,end_type,winner,insert_date,update_date,version)  values(?,?,?,?,?,?,?,?,?,?,?,?,?)")
	if r.Level == DEBUG {
		log.Println(SQL_INSERT, exeSql)
	}
	_, err := r.DB.Exec(exeSql, p.AutoId, p.BatchNo, p.PlayName, p.FromPlayer, p.ToPlayer, p.CoinCnt, p.MedalCnt, p.Status, p.EndType, p.Winner, p.InsertDate, p.UpdateDate, p.Version)
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

func (r AfPlayList) InsertEntity(p AfPlay, tr *sql.Tx) error {
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

	if p.PlayName != "" {
		colNames += "play_name,"
		colTags += "?,"
		valSlice = append(valSlice, p.PlayName)
	}

	if p.FromPlayer != "" {
		colNames += "from_player,"
		colTags += "?,"
		valSlice = append(valSlice, p.FromPlayer)
	}

	if p.ToPlayer != "" {
		colNames += "to_player,"
		colTags += "?,"
		valSlice = append(valSlice, p.ToPlayer)
	}

	if p.CoinCnt != 0 {
		colNames += "coin_cnt,"
		colTags += "?,"
		valSlice = append(valSlice, p.CoinCnt)
	}

	if p.MedalCnt != 0 {
		colNames += "medal_cnt,"
		colTags += "?,"
		valSlice = append(valSlice, p.MedalCnt)
	}

	if p.Status != "" {
		colNames += "status,"
		colTags += "?,"
		valSlice = append(valSlice, p.Status)
	}

	if p.EndType != "" {
		colNames += "end_type,"
		colTags += "?,"
		valSlice = append(valSlice, p.EndType)
	}

	if p.Winner != "" {
		colNames += "winner,"
		colTags += "?,"
		valSlice = append(valSlice, p.Winner)
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

	if p.Version != 0 {
		colNames += "version,"
		colTags += "?,"
		valSlice = append(valSlice, p.Version)
	}

	colNames = strings.TrimRight(colNames, ",")
	colTags = strings.TrimRight(colTags, ",")
	exeSql := fmt.Sprintf("Insert into  af_play(%s)  values(%s)", colNames, colTags)
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

func (r AfPlayList) InsertMap(m map[string]interface{}, tr *sql.Tx) error {
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

	exeSql := fmt.Sprintf("Insert into  af_play(%s)  values(%s)", colNames, colTags)
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

func (r AfPlayList) UpdataEntity(keyNo string, p AfPlay, tr *sql.Tx) error {
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

	if p.PlayName != "" {
		colNames += "play_name=?,"

		valSlice = append(valSlice, p.PlayName)
	}

	if p.FromPlayer != "" {
		colNames += "from_player=?,"

		valSlice = append(valSlice, p.FromPlayer)
	}

	if p.ToPlayer != "" {
		colNames += "to_player=?,"

		valSlice = append(valSlice, p.ToPlayer)
	}

	if p.CoinCnt != 0 {
		colNames += "coin_cnt=?,"
		valSlice = append(valSlice, p.CoinCnt)
	}

	if p.MedalCnt != 0 {
		colNames += "medal_cnt=?,"
		valSlice = append(valSlice, p.MedalCnt)
	}

	if p.Status != "" {
		colNames += "status=?,"

		valSlice = append(valSlice, p.Status)
	}

	if p.EndType != "" {
		colNames += "end_type=?,"

		valSlice = append(valSlice, p.EndType)
	}

	if p.Winner != "" {
		colNames += "winner=?,"

		valSlice = append(valSlice, p.Winner)
	}

	if p.InsertDate != 0 {
		colNames += "insert_date=?,"
		valSlice = append(valSlice, p.InsertDate)
	}

	if p.UpdateDate != 0 {
		colNames += "update_date=?,"
		valSlice = append(valSlice, p.UpdateDate)
	}

	if p.Version != 0 {
		colNames += "version=?,"
		valSlice = append(valSlice, p.Version)
	}

	colNames = strings.TrimRight(colNames, ",")
	valSlice = append(valSlice, keyNo)

	exeSql := fmt.Sprintf("update  af_play  set %s  where auto_id=? ", colNames)
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

func (r AfPlayList) UpdateMap(keyNo string, m map[string]interface{}, tr *sql.Tx) error {
	l := time.Now()

	var colNames string
	valSlice := make([]interface{}, 0)
	for k, v := range m {
		colNames += k + "=?,"
		valSlice = append(valSlice, v)
	}
	valSlice = append(valSlice, keyNo)
	colNames = strings.TrimRight(colNames, ",")
	updateSql := fmt.Sprintf("Update af_play set %s where auto_id=?", colNames)
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

func (r AfPlayList) Delete(keyNo string, tr *sql.Tx) error {
	l := time.Now()
	delSql := fmt.Sprintf("Delete from  af_play  where auto_id=?")
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
