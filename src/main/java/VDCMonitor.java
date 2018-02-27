
import com.mashape.unirest.http.Unirest;
import com.mashape.unirest.http.exceptions.UnirestException;
import com.mashape.unirest.request.HttpRequestWithBody;
import de.tub.ditas.job.Heartbeat;
import de.tub.ditas.job.IPTrafNg;
import de.tub.ditas.MonitorConfig;
import org.pmw.tinylog.Configurator;
import org.pmw.tinylog.Level;
import org.pmw.tinylog.Logger;
import org.pmw.tinylog.writers.ConsoleWriter;

import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;

public class VDCMonitor {
    public static void main(String[] args) {
        Configurator.defaultConfig().writer(new ConsoleWriter()).level((args.length > 0 ? Level.TRACE : Level.OFF)).activate();
        MonitorConfig config = MonitorConfig.fromEnvironment();

        waitForConnection(config);
        sendMappings(config);
        ScheduledExecutorService jobRunner = Executors.newScheduledThreadPool(2);


        jobRunner.scheduleAtFixedRate(new Heartbeat(config), 5, config.invocationInverval, TimeUnit.MILLISECONDS);
        jobRunner.scheduleAtFixedRate(new IPTrafNg(config), 5, config.invocationInverval, TimeUnit.MILLISECONDS);
        Logger.info("started VDC Monitor");
    }

    private static void waitForConnection(MonitorConfig config) {
        while (!connectToElastic(config)) {
            Logger.info("waiting for elasticsearch connection");
            try {
                Thread.sleep(5000);
            } catch (InterruptedException e) {
            }
        }
    }

    private static boolean connectToElastic(MonitorConfig config) {
        try {
            Unirest.head(config.getElasticSearchURL()).asString();
            return true;
        } catch (UnirestException e) {
            return false;
        }
    }

    private static void sendMappings(MonitorConfig config) {
        HttpRequestWithBody requestHeart = Unirest.put(config.getElasticSearchURL() + "traffic" + "/_mapping/_doc");
        requestHeart.body("\n" +
                "{\n" +
                "    \"settings\":{\n" +
                "        \"number_of_shards\":1,\n" +
                "        \"number_of_replicas\":0\n" +
                "    },\n" +
                "    \"mappings\":{\n" +
                "        \"traffic\":{\n" +
                "            \"properties\":{\n" +
                "                \"@timestamp\":{\n" +
                "                    \"type\":\"date\"\n" +
                "                },\n" +
                "                \"traffic.component\":{\n" +
                "                    \"type\":\"text\"\n" +
                "                },\n" +
                "                \"traffic.bytes\":{\n" +
                "                    \"type\":\"long\"\n" +
                "                },\n" +
                "                \"traffic.date\":{\n" +
                "                    \"type\":\"text\"\n" +
                "                }   \n" +
                "                \n" +
                "            }\n" +
                "        },\n" +
                "        \"heartbeat\":{\n" +
                "            \"properties\":{\n" +
                "                \"@timestamp\":{\n" +
                "                    \"type\":\"date\"\n" +
                "                },\n" +
                "                \"traffic.interval\":{\n" +
                "                    \"type\":\"long\"\n" +
                "                },\n" +
                "                \"traffic.time\":{\n" +
                "                    \"type\":\"long\"\n" +
                "                }\n" +
                "\n" +
                "            }\n" +
                "        }\n" +
                "    }\n" +
                "}");
        try {
            Logger.debug("Send mapping got " +requestHeart.asJson());
        } catch (UnirestException e) {
            e.printStackTrace();
        }
    }
}
