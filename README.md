# table-rock-lake-temp

Docker Image to fetch lake temperature at Table Rock Lake and send it to MQTT or Influxdb

## Environment Variables

The following environment variables are used to configure the container:

-   INFLUXDB_SERVER: the url of the influxdb server
-   INFLUXDB_PORT: default 8086, port used to connect to influxdb server
-   INFLUXDB_USERNAME: default "", username for connecting to the influxdb server
-   INFLUXDB_PASSWORD: default "", password for connecting to the influxdb server
-   INFLUXDB_DATABASE: default "lakeinfo", database to use for storing measurements
-   INFLUXDB_PREFIX: default "", prefix (usually the lake name separated by underscores) for the measurements stored in the influxdb server
-   MQTT_SERVER: the url of the MQTT server
-   MQTT_PORT: default 1883, port used to connect to MQTT server
-   MQTT_USERNAME: default "", username for connecting to the MQTT server
-   MQTT_PASSWORD: default "", password for connecting to the MQTT server
-   MQTT_PREFIX: default "", prefix (usually the lake name separated by underscores) for the measurements stored in the MQTT server
