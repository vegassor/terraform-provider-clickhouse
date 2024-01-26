package clickhouse_client

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/vegassor/terraform-provider-clickhouse/internal/mock"
	"testing"
)

func TestDataBaseSQL(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	conn := mock.NewMockConn(mockCtrl)
	rows := mock.NewMockRows(mockCtrl)
	rows.EXPECT().Next().Return(true).Times(1)
	rows.EXPECT().Scan(gomock.Any()).Return(nil).Times(1)

	expectedQuery := `SELECT "name", "auth_type", "host_ip", "host_names", "host_names_regexp", "host_names_like"
		FROM "system"."users"
		WHERE "name" = 'my_user'`
	conn.EXPECT().Query(ctx, expectedQuery).Return(rows, nil).Times(1)

	client := ClickHouseClient{Conn: conn}
	client.GetUser(ctx, "my_user")
}
