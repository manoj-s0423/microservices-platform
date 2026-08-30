package com.shopstream.user.service;

import com.shopstream.user.dto.*;
import com.shopstream.user.exception.EmailAlreadyExistsException;
import com.shopstream.user.exception.InvalidCredentialsException;
import com.shopstream.user.exception.UserNotFoundException;
import com.shopstream.user.model.User;
import com.shopstream.user.repository.UserRepository;
import com.shopstream.user.security.JwtUtil;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.UUID;

@Service
public class UserService {

    private static final Logger log = LoggerFactory.getLogger(UserService.class);

    private final UserRepository userRepository;
    private final PasswordEncoder passwordEncoder;
    private final JwtUtil jwtUtil;

    public UserService(UserRepository userRepository, PasswordEncoder passwordEncoder, JwtUtil jwtUtil) {
        this.userRepository = userRepository;
        this.passwordEncoder = passwordEncoder;
        this.jwtUtil = jwtUtil;
    }

    @Transactional
    public AuthResponse register(RegisterRequest request) {
        if (userRepository.existsByEmail(request.email())) {
            throw new EmailAlreadyExistsException(request.email());
        }

        User user = new User();
        user.setEmail(request.email());
        user.setPasswordHash(passwordEncoder.encode(request.password()));
        user.setFirstName(request.firstName());
        user.setLastName(request.lastName());
        user.setRole(User.Role.CUSTOMER);
        user.setStatus(User.Status.ACTIVE);

        User saved = userRepository.save(user);
        log.info("registered new user id={}", saved.getId());

        String token = jwtUtil.generateToken(saved.getId().toString(), saved.getRole().name());
        return AuthResponse.of(token, jwtUtil.getExpirationMs(), UserResponse.from(saved));
    }

    @Transactional(readOnly = true)
    public AuthResponse login(LoginRequest request) {
        User user = userRepository.findByEmail(request.email())
                .orElseThrow(InvalidCredentialsException::new);

        if (!passwordEncoder.matches(request.password(), user.getPasswordHash())) {
            throw new InvalidCredentialsException();
        }

        if (user.getStatus() != User.Status.ACTIVE) {
            log.warn("login attempt on non-active account id={} status={}", user.getId(), user.getStatus());
            throw new InvalidCredentialsException();
        }

        String token = jwtUtil.generateToken(user.getId().toString(), user.getRole().name());
        return AuthResponse.of(token, jwtUtil.getExpirationMs(), UserResponse.from(user));
    }

    @Transactional(readOnly = true)
    public UserResponse getById(UUID id) {
        User user = userRepository.findById(id).orElseThrow(() -> new UserNotFoundException(id));
        return UserResponse.from(user);
    }
}
