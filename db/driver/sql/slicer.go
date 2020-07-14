package sql

import (
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/google/uuid"

	"qlova.store/db"
)

//Slicer implements db.Slicer
type slicer struct {
	length int

	columns []db.Column

	values []db.Value

	Query
}

func (slice slicer) Columns(col db.Column, cols ...db.Column) db.Slicer {
	slice.columns = append(cols, col)
	return slice
}

//MarshalJSON implements json.Marshaler
func (slice slicer) MarshalJSON() ([]byte, error) {
	var q = slice.Query

	var values = slice.values

	var head Query
	head.WriteString(`SELECT `)
	head.WriteColumn(values[0].GetColumn())
	for _, col := range values[1:] {
		head.WriteByte(',')
		head.WriteColumn(col.GetColumn())
	}
	head.WriteString(` FROM `)
	head.WriteString(q.Table.Name)
	head.WriteString(` `)

	head.WriteByte(' ')
	head.WriteQuery(&q)

	rows, err := q.Driver.DB.QueryContext(q.Driver.Context, head.String(), head.Values...)
	if err != nil {
		return nil, head.Error(err)
	}

	results := make([]interface{}, len(values))

	pointers := make([]interface{}, len(values))
	for i := range values {
		pointers[i] = &results[i]
	}

	var buffer bytes.Buffer
	buffer.WriteString(`[`)

	var index int
	for index = 0; rows.Next(); index++ {
		if index != 0 {
			buffer.WriteByte(',')
		}

		if err := rows.Err(); err != nil {
			return nil, head.Error(err)
		}

		if err := rows.Scan(pointers...); err != nil {
			return nil, head.Error(err)
		}

		buffer.WriteString(`{`)
		for i, result := range results {
			buffer.WriteString(strconv.Quote(values[i].GetColumn().Name))
			buffer.WriteByte(':')

			switch values[i].(type) {
			case db.UUID:
				var id uuid.UUID
				id.Scan(result)
				buffer.WriteString(strconv.Quote(id.String()))
			default:
				encoded, err := json.Marshal(result)
				if err != nil {
					return nil, head.Error(err)
				}
				buffer.Write(encoded)
			}

			if i < len(results)-1 {
				buffer.WriteByte(',')
			}
		}
		buffer.WriteString(`}`)
	}

	buffer.WriteByte(']')

	return buffer.Bytes(), nil
}

//Into implements db.Slicer.Into
func (slice slicer) Into(model db.Connectable, extras ...db.Connectable) error {
	var q = slice.Query

	var columns = slice.columns

	if columns == nil {
		columns = db.Columns(model)
	}

	var head Query
	head.WriteString(`SELECT `)
	head.WriteColumn(columns[0])
	for _, col := range columns[1:] {
		head.WriteByte(',')
		head.WriteColumn(col)
	}
	head.WriteString(` FROM `)
	head.WriteString(q.Table.Name)
	head.WriteString(` `)

	head.WriteByte(' ')
	head.WriteQuery(&q)

	rows, err := q.Driver.DB.QueryContext(q.Driver.Context, head.String(), head.Values...)
	if err != nil {
		return head.Error(err)
	}

	var slices = db.Slices(model, slice.length, columns...)
	var pointers = make([]interface{}, len(slices))

	var index int
	for index = 0; rows.Next(); index++ {
		if err := rows.Err(); err != nil {
			return head.Error(err)
		}

		for i := range pointers {
			pointers[i] = slices[i].Index(index).Addr().Interface()
		}

		if err := rows.Scan(pointers...); err != nil {
			return head.Error(err)
		}
	}

	model.Slice(0, index)

	return nil
}
