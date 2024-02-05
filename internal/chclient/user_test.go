package chclient

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/vegassor/terraform-provider-clickhouse/internal/mock"
	"testing"
)

func TestGetUserSQL(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	conn := mock_driver.NewMockConn(mockCtrl)
	rows := mock_driver.NewMockRows(mockCtrl)
	rows.EXPECT().Next().Return(true).Times(1)
	rows.EXPECT().Scan(gomock.Any()).Return(nil).Times(1)

	expectedQuery := `SELECT "name", "auth_type", "host_ip", "host_names",
"host_names_regexp", "host_names_like", "default_database"
FROM "system"."users"
WHERE "name" = 'my_user'`
	conn.EXPECT().Query(ctx, expectedQuery).Return(rows, nil).Times(1)

	client := ClickHouseClient{Conn: conn}
	_, _ = client.GetUser(ctx, "my_user")
}
