package clickhouse_client

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"net"
	"strings"
)

type ClickHouseUserAuthType interface {
	getIdentifiedWithQuery() string
}

type Sha256HashAuth struct {
	Hash string
	Salt string
}

func (auth Sha256HashAuth) getIdentifiedWithQuery() string {
	result := "sha256_hash BY " + QuoteValue(auth.Hash)
	if auth.Salt != "" {
		result += " SALT " + QuoteValue(auth.Salt)
	}
	return result
}

type Sha256PasswordAuth struct {
	Password string
}

func (auth Sha256PasswordAuth) getIdentifiedWithQuery() string {
	return "sha256_password BY " + QuoteValue(auth.Password)
}

type ClickHouseUserHosts struct {
	Ip     []net.IPNet
	Name   []string
	Regexp []string
	Like   []string
}

func (hosts *ClickHouseUserHosts) GetHostsQuery() string {
	if hosts == nil {
		return "ANY"
	}

	hostsCount := len(hosts.Ip) + len(hosts.Name) + len(hosts.Regexp) + len(hosts.Like)
	if hostsCount == 0 {
		return "NONE"
	}

	result := make([]string, 0, hostsCount)

	for _, fqdn := range hosts.Name {
		result = append(result, "NAME "+QuoteValue(fqdn))
	}

	for _, regexp := range hosts.Regexp {
		result = append(result, "REGEXP "+QuoteValue(regexp))
	}

	for _, like := range hosts.Like {
		result = append(result, "LIKE "+QuoteValue(like))
	}

	for _, like := range hosts.Ip {
		result = append(result, "IP "+QuoteValue(like.String()))
	}

	return strings.Join(result, ", ")
}

type DefaultDatabase string

func (db DefaultDatabase) String() string {
	if db == "" {
		return "NONE"
	}

	return string(db)
}

type ClickHouseUser struct {
	Name            string
	Auth            ClickHouseUserAuthType
	Hosts           *ClickHouseUserHosts
	DefaultDatabase DefaultDatabase
}

func (client *ClickHouseClient) CreateUser(ctx context.Context, user ClickHouseUser) error {
	query := fmt.Sprintf(
		"CREATE USER %s IDENTIFIED WITH %s HOST %s DEFAULT DATABASE %s",
		user.Name,
		user.Auth.getIdentifiedWithQuery(),
		user.Hosts.GetHostsQuery(),
		user.DefaultDatabase.String(),
	)

	tflog.Info(ctx, "Creating a user", map[string]interface{}{"query": query})

	return client.Conn.Exec(ctx, query)
}

func (client *ClickHouseClient) DropUser(ctx context.Context, user string) error {
	query := fmt.Sprintf(
		"DROP USER %s",
		QuoteValue(user),
	)

	tflog.Info(ctx, "Dropping a user", map[string]interface{}{"query": query})

	return client.Conn.Exec(ctx, query)
}

func (client *ClickHouseClient) GetUser(ctx context.Context, name string) (ClickHouseUser, error) {
	query := fmt.Sprintf(
		`SELECT "name", "auth_type", "host_ip", "host_names",
"host_names_regexp", "host_names_like", "default_database"
FROM "system"."users"
WHERE "name" = %s`,
		QuoteValue(name),
	)

	tflog.Info(ctx, "Querying a user", map[string]interface{}{"query": query})

	rows, err := client.Conn.Query(ctx, query)
	if err != nil {
		return ClickHouseUser{}, err
	}

	if !rows.Next() {
		return ClickHouseUser{}, fmt.Errorf(
			"could not find user %s: no records returned. Query: `%s`",
			name,
			query,
		)
	}

	var nameReceived string
	var authType string
	var chUserHosts = &ClickHouseUserHosts{}
	var chUserHostsIp []string
	var defaultDb string

	err = rows.Scan(
		&nameReceived,
		&authType,
		&chUserHostsIp,
		&chUserHosts.Name,
		&chUserHosts.Regexp,
		&chUserHosts.Like,
		&defaultDb,
	)
	if err != nil {
		return ClickHouseUser{}, err
	}

	for _, ip := range chUserHostsIp {
		_, ipNet, err := net.ParseCIDR(ip)
		if err != nil {
			parsedIp := net.ParseIP(ip)
			if parsedIp == nil {
				return ClickHouseUser{}, err
			}
			_, ipNet, _ = net.ParseCIDR(parsedIp.String() + "/32")
		}
		chUserHosts.Ip = append(chUserHosts.Ip, *ipNet)
	}

	if len(chUserHosts.Name) == 0 &&
		len(chUserHosts.Like) == 0 &&
		len(chUserHosts.Regexp) == 0 &&
		len(chUserHosts.Ip) == 1 {
		ones, _ := chUserHosts.Ip[0].Mask.Size()
		if ones == 0 {
			chUserHosts = nil
		}
	}

	return ClickHouseUser{
		Name:            nameReceived,
		Auth:            Sha256PasswordAuth{},
		Hosts:           chUserHosts,
		DefaultDatabase: DefaultDatabase(defaultDb),
	}, nil
}

func (client *ClickHouseClient) AlterUser(ctx context.Context, origName string, user ClickHouseUser) error {
	shouldRaname := origName != user.Name
	var renameQuery string
	if shouldRaname {
		renameQuery = "RENAME TO " + QuoteValue(user.Name)
	}

	query := fmt.Sprintf(
		"ALTER USER %s %s IDENTIFIED WITH %s HOST %s DEFAULT DATABASE %s",
		origName,
		renameQuery,
		user.Auth.getIdentifiedWithQuery(),
		user.Hosts.GetHostsQuery(),
		user.DefaultDatabase.String(),
	)

	tflog.Info(ctx, "Altering a user", map[string]interface{}{
		"query":        query,
		"origUsername": origName,
		"newUsername":  user.Name,
		"rename":       shouldRaname,
	})

	return client.Conn.Exec(ctx, query)
}
