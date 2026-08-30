package com.shopstream.user;

import org.junit.jupiter.api.Test;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.ActiveProfiles;

/**
 * Smoke test: the Spring context (including JWT key init, Flyway-free H2
 * schema, and security filter chain) must load successfully.
 */
@SpringBootTest
@ActiveProfiles("test")
class UserServiceApplicationTests {

    @Test
    void contextLoads() {
    }
}
