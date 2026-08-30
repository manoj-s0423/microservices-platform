package com.shopstream.user.service;

import com.shopstream.user.dto.LoginRequest;
import com.shopstream.user.dto.RegisterRequest;
import com.shopstream.user.exception.EmailAlreadyExistsException;
import com.shopstream.user.exception.InvalidCredentialsException;
import com.shopstream.user.model.User;
import com.shopstream.user.repository.UserRepository;
import com.shopstream.user.security.JwtUtil;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;

import java.util.Optional;
import java.util.UUID;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class UserServiceTest {

    @Mock
    private UserRepository userRepository;

    @Mock
    private JwtUtil jwtUtil;

    private PasswordEncoder passwordEncoder;

    @InjectMocks
    private UserService userService;

    @BeforeEach
    void setUp() {
        passwordEncoder = new BCryptPasswordEncoder();
        userService = new UserService(userRepository, passwordEncoder, jwtUtil);
    }

    @Test
    void register_savesNewUser_andReturnsToken() {
        RegisterRequest request = new RegisterRequest("new@shopstream.dev", "Password123!", "New", "User");
        when(userRepository.existsByEmail(request.email())).thenReturn(false);
        when(userRepository.save(any(User.class))).thenAnswer(inv -> {
            User u = inv.getArgument(0);
            u.setId(UUID.randomUUID());
            return u;
        });
        when(jwtUtil.generateToken(any(), any())).thenReturn("fake-jwt-token");
        when(jwtUtil.getExpirationMs()).thenReturn(3600000L);

        var response = userService.register(request);

        assertThat(response.accessToken()).isEqualTo("fake-jwt-token");
        assertThat(response.user().email()).isEqualTo("new@shopstream.dev");
        verify(userRepository).save(any(User.class));
    }

    @Test
    void register_rejectsDuplicateEmail() {
        RegisterRequest request = new RegisterRequest("existing@shopstream.dev", "Password123!", "A", "B");
        when(userRepository.existsByEmail(request.email())).thenReturn(true);

        assertThatThrownBy(() -> userService.register(request))
                .isInstanceOf(EmailAlreadyExistsException.class);

        verify(userRepository, never()).save(any());
    }

    @Test
    void login_withCorrectPassword_returnsToken() {
        User user = new User();
        user.setId(UUID.randomUUID());
        user.setEmail("jane@shopstream.dev");
        user.setPasswordHash(passwordEncoder.encode("correct-password"));
        user.setRole(User.Role.CUSTOMER);
        user.setStatus(User.Status.ACTIVE);

        when(userRepository.findByEmail("jane@shopstream.dev")).thenReturn(Optional.of(user));
        when(jwtUtil.generateToken(any(), any())).thenReturn("fake-jwt-token");
        when(jwtUtil.getExpirationMs()).thenReturn(3600000L);

        var response = userService.login(new LoginRequest("jane@shopstream.dev", "correct-password"));

        assertThat(response.accessToken()).isEqualTo("fake-jwt-token");
    }

    @Test
    void login_withWrongPassword_throwsInvalidCredentials() {
        User user = new User();
        user.setId(UUID.randomUUID());
        user.setEmail("jane@shopstream.dev");
        user.setPasswordHash(passwordEncoder.encode("correct-password"));
        user.setStatus(User.Status.ACTIVE);

        when(userRepository.findByEmail("jane@shopstream.dev")).thenReturn(Optional.of(user));

        assertThatThrownBy(() -> userService.login(new LoginRequest("jane@shopstream.dev", "wrong-password")))
                .isInstanceOf(InvalidCredentialsException.class);
    }

    @Test
    void login_withUnknownEmail_throwsInvalidCredentials() {
        when(userRepository.findByEmail("nobody@shopstream.dev")).thenReturn(Optional.empty());

        assertThatThrownBy(() -> userService.login(new LoginRequest("nobody@shopstream.dev", "whatever")))
                .isInstanceOf(InvalidCredentialsException.class);
    }

    @Test
    void login_withSuspendedAccount_throwsInvalidCredentials() {
        User user = new User();
        user.setId(UUID.randomUUID());
        user.setEmail("suspended@shopstream.dev");
        user.setPasswordHash(passwordEncoder.encode("correct-password"));
        user.setStatus(User.Status.SUSPENDED);

        when(userRepository.findByEmail("suspended@shopstream.dev")).thenReturn(Optional.of(user));

        assertThatThrownBy(() -> userService.login(new LoginRequest("suspended@shopstream.dev", "correct-password")))
                .isInstanceOf(InvalidCredentialsException.class);
    }
}
