package impl

import (
	"time"

	"gitlab.kenda.com.tw/kenda/mcom/impl/orm/models"
)

func (session *session) listWorkOrders(departmentID string, reversedDate time.Time) ([]models.WorkOrder, error) {
	var results []models.WorkOrder
	err := session.db.
		Where(` "department_id" = ? AND "reserved_date" = ? `, departmentID, reversedDate.Format("2006-01-02")).
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}
