package com.example.echo;

import static org.junit.jupiter.api.Assertions.*;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.HttpURLConnection;
import java.net.URI;
import java.net.URL;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.time.Duration;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.TimeUnit;

import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.DisplayName;

import com.sun.net.httpserver.HttpServer;

public class EchoApplicationTest {
    private static HttpServer server;
    private static final int PORT = 8083;
    private static final HttpClient client = HttpClient.newBuilder()
            .version(HttpClient.Version.HTTP_1_1)
            .connectTimeout(Duration.ofSeconds(10))
            .build();

    @BeforeAll
    static void setup() throws IOException {
        CompletableFuture.runAsync(() -> {
            try {
                EchoApplication.main(new String[]{});
            } catch (IOException e) {
                fail("Server failed to start: " + e.getMessage());
            }
        });
        
        // Wait for server to start
        Thread.sleep(1000);
    }

    @Test
    @DisplayName("Test server starts successfully")
    void testMain() {
        assertDoesNotThrow(() -> {
            URL healthUrl = new URI("http://localhost:" + PORT + "/health").toURL();
            HttpURLConnection conn = (HttpURLConnection) healthUrl.openConnection();
            assertEquals(200, conn.getResponseCode());
        });
    }

    @Test
    @DisplayName("Test echo endpoint with simple JSON")
    void testEchoSimpleJson() throws Exception {
        String json = "{\"message\":\"test\"}";
        HttpRequest request = HttpRequest.newBuilder()
                .uri(new URI("http://localhost:" + PORT + "/"))
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(json))
                .build();

        HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());
        
        assertEquals(200, response.statusCode());
        assertEquals(json, response.body());
        assertEquals("application/json", response.headers().firstValue("Content-Type").orElse(null));
    }

    @Test
    @DisplayName("Test echo endpoint with complex JSON")
    void testEchoComplexJson() throws Exception {
        String json = "{\"string\":\"value\",\"number\":42,\"boolean\":true,\"array\":[1,2,3],\"nested\":{\"key\":\"value\"}}";
        HttpRequest request = HttpRequest.newBuilder()
                .uri(new URI("http://localhost:" + PORT + "/"))
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(json))
                .build();

        HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());
        
        assertEquals(200, response.statusCode());
        assertEquals(json, response.body());
    }

    @Test
    @DisplayName("Test health check endpoint")
    void testHealthCheck() throws Exception {
        HttpRequest request = HttpRequest.newBuilder()
                .uri(new URI("http://localhost:" + PORT + "/health"))
                .GET()
                .build();

        HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());
        
        assertEquals(200, response.statusCode());
        assertEquals("{\"status\":\"ok\"}", response.body());
    }

    @Test
    @DisplayName("Test method not allowed")
    void testMethodNotAllowed() throws Exception {
        HttpRequest request = HttpRequest.newBuilder()
                .uri(new URI("http://localhost:" + PORT + "/"))
                .GET()
                .build();

        HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());
        assertEquals(405, response.statusCode());
    }

    @Test
    @DisplayName("Test health endpoint rejects POST")
    void testHealthPostNotAllowed() throws Exception {
        HttpRequest request = HttpRequest.newBuilder()
                .uri(new URI("http://localhost:" + PORT + "/health"))
                .POST(HttpRequest.BodyPublishers.ofString("{}"))
                .build();

        HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());
        assertEquals(405, response.statusCode());
    }

    @Test
    @DisplayName("Test concurrent requests")
    void testConcurrentRequests() throws Exception {
        int numRequests = 10;
        CompletableFuture<HttpResponse<String>>[] futures = new CompletableFuture[numRequests];
        
        for (int i = 0; i < numRequests; i++) {
            String json = String.format("{\"request\":%d}", i);
            HttpRequest request = HttpRequest.newBuilder()
                    .uri(new URI("http://localhost:" + PORT + "/"))
                    .header("Content-Type", "application/json")
                    .POST(HttpRequest.BodyPublishers.ofString(json))
                    .build();
                    
            futures[i] = client.sendAsync(request, HttpResponse.BodyHandlers.ofString());
        }
        
        CompletableFuture.allOf(futures).get(5, TimeUnit.SECONDS);
        
        for (int i = 0; i < numRequests; i++) {
            HttpResponse<String> response = futures[i].get();
            assertEquals(200, response.statusCode());
            assertEquals(String.format("{\"request\":%d}", i), response.body());
        }
    }
}
