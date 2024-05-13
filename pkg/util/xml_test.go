package util

import "testing"

func TestMarshal(t *testing.T) {

	tests := []struct {
		name string
		x    *XMLConfiguration
		want string
	}{
		{
			name: "empty",
			x:    &XMLConfiguration{},
			want: `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type="text/xsl" href="configuration.xsl"?>
<configuration></configuration>`,
		},
		{
			name: "one",
			x: &XMLConfiguration{
				Properties: []*Property{
					{
						Name:        "name",
						Value:       "value",
						Description: "description",
					},
				},
			},
			want: `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type="text/xsl" href="configuration.xsl"?>
<configuration>
    <property>
        <name>name</name>
        <value>value</value>
        <description>description</description>
    </property>
</configuration>`,
		},
		{
			name: "two",
			x: &XMLConfiguration{
				Properties: []*Property{
					{
						Name:  "name1",
						Value: "value1",
					},
				},
			},
			want: `<?xml version="1.0" encoding="UTF-8"?>
<?xml-stylesheet type="text/xsl" href="configuration.xsl"?>
<configuration>
    <property>
        <name>name1</name>
        <value>value1</value>
    </property>
</configuration>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.x.Marshal()
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Marshal() got = %v, want %v", got, tt.want)
			}
		})
	}
}
