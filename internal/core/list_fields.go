package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// maxListFields bounds the columns a list adds to each of its rows: a page
// keeps its size near that of the summary, a question naming the few
// properties it needs.
const maxListFields = 8

// NodeListFields are the properties list_nodes adds to each node on demand:
// a property of many nodes, such as their team or city, is read in one page
// rather than a get_node each. Named one by one, so that a property get_node
// reads later is not listed raw without a decision; get_node reads them all,
// the nodes coming from one route, /nodes, in one form: a number not set is
// 0, a date is in collector local time. mem_bytes is left out: the collector
// stores MiB under it, which get_node gives as mem_mb.
var NodeListFields = []string{
	"fqdn", "role", "type", "snooze_till", "updated",
	"os_name", "os_vendor", "os_release", "os_kernel", "os_arch", "tz", "last_boot",
	"manufacturer", "model", "serial", "cpu_model", "cpu_vendor", "cpu_cores", "cpu_threads", "cpu_dies", "cpu_freq",
	"mem_banks", "mem_slots", "bios_version", "sp_version",
	"team_responsible", "team_integ", "team_support",
	"loc_country", "loc_city", "loc_addr", "loc_zip", "loc_building", "loc_floor", "loc_room", "loc_rack",
	"warranty_end", "maintenance_end", "hw_obs_warn_date", "hw_obs_alert_date", "os_obs_warn_date", "os_obs_alert_date",
}

// ServiceListFields are the properties list_services adds to each service on
// demand: texts alone, which the collector gives in one form whether the
// services are read from /services or, for those of a node, joined from
// /services_instances. Its dates, numbers and rates are left out: joined,
// a date and a number not set come in another form, and get_service gives
// the rates in percent and the high availability as a boolean.
var ServiceListFields = []string{"svc_comment", "svc_autostart", "svc_drpnode", "svc_drpnodes", "svc_drptype"}

// checkListFields keeps the fields asked, once each in their order: a
// property of the summary, already in each row, is left out, and one out of
// allowed, or more than maxListFields, is refused.
func checkListFields(fields, allowed, summary []string) ([]string, error) {
	var out []string
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if slices.Contains(out, field) || slices.Contains(summary, field) {
			continue
		}
		if !slices.Contains(allowed, field) {
			return nil, fmt.Errorf("unknown field %q: the fields are %s; the columns of the summary are in each row already", field, strings.Join(allowed, ", "))
		}
		out = append(out, field)
	}
	if len(out) > maxListFields {
		return nil, fmt.Errorf("%d fields asked, at most %d", len(out), maxListFields)
	}
	return out, nil
}

// collectorRow is a row of a collector list decoded into its typed form T,
// keeping its raw form for the fields asked: decoded inside the read of the
// answer, a row of another form is an incompatible answer.
type collectorRow[T any] struct {
	Typed T
	Raw   json.RawMessage
}

func (r *collectorRow[T]) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &r.Typed); err != nil {
		return err
	}
	r.Raw = append(json.RawMessage(nil), data...)
	return nil
}

// rowFields gives the fields of a collector row, its keys prefixed as the
// collector names joined columns; a field not set, null or empty, is left
// out.
func rowFields(row json.RawMessage, fields []string, prefix string) (map[string]any, error) {
	if len(fields) == 0 {
		return nil, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(row))
	decoder.UseNumber()
	var values map[string]any
	if err := decoder.Decode(&values); err != nil {
		return nil, fmt.Errorf("decode the fields of a row: %w", err)
	}
	out := make(map[string]any, len(fields))
	for _, field := range fields {
		value, ok := values[prefix+field]
		if !ok || value == nil || value == "" {
			continue
		}
		out[field] = value
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}
