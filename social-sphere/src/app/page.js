"use client";

import Link from "next/link";
import Image from "next/image";
import { motion } from "motion/react";
import {
    Users,
    MessageCircle,
    Bell,
    Globe,
    Shield,
    ArrowRight
} from "lucide-react";

// Animation variants
const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
        opacity: 1,
        transition: {
            staggerChildren: 0.12,
            delayChildren: 0.3,
        }
    }
};

const itemVariants = {
    hidden: { opacity: 0, y: 40, filter: 'blur(8px)' },
    visible: {
        opacity: 1,
        y: 0,
        filter: 'blur(0px)',
        transition: { duration: 0.7, ease: [0.25, 0.1, 0, 1] }
    }
};

const featureCardVariants = {
    hidden: { opacity: 0, y: 60, scale: 0.95 },
    visible: {
        opacity: 1,
        y: 0,
        scale: 1,
        transition: { duration: 0.6, ease: [0.25, 0.1, 0, 1] }
    }
};

// Hero Sphere Component with floating particles
function HeroSphere() {
    // Orbital particles around the sphere
    const particles = [
        { size: 6, radius: 200, duration: 20, delay: 0, opacity: 0.8 },
        { size: 4, radius: 220, duration: 25, delay: 5, opacity: 0.6 },
        { size: 5, radius: 180, duration: 18, delay: 2, opacity: 0.7 },
        { size: 3, radius: 240, duration: 30, delay: 8, opacity: 0.5 },
        { size: 4, radius: 160, duration: 15, delay: 3, opacity: 0.7 },
        { size: 5, radius: 260, duration: 35, delay: 12, opacity: 0.4 },
    ];

    return (
        <div className="relative w-[500px] h-[500px] flex items-center justify-center">
            {/* Multi-layered glow effect behind sphere */}
            <div className="absolute w-[350px] h-[350px] rounded-full bg-linear-to-br from-purple-500/30 via-pink-500/20 to-orange-500/30 blur-[80px]" />
            <div className="absolute w-[280px] h-[280px] rounded-full bg-linear-to-tr from-blue-500/25 via-violet-500/20 to-pink-500/25 blur-[60px]" />

            {/* Orbital rings */}
            <motion.div
                className="absolute w-[400px] h-[400px] rounded-full border border-foreground/5"
                animate={{ rotate: 360 }}
                transition={{ duration: 60, repeat: Infinity, ease: "linear" }}
            />
            <motion.div
                className="absolute w-[320px] h-80 rounded-full border border-foreground/8"
                animate={{ rotate: -360 }}
                transition={{ duration: 45, repeat: Infinity, ease: "linear" }}
            />

            {/* Orbiting particles */}
            {particles.map((particle, i) => (
                <motion.div
                    key={i}
                    className="absolute"
                    style={{
                        width: particle.radius * 2,
                        height: particle.radius * 2,
                    }}
                    animate={{ rotate: 360 }}
                    transition={{
                        duration: particle.duration,
                        repeat: Infinity,
                        ease: "linear",
                        delay: particle.delay,
                    }}
                >
                    <motion.div
                        className="absolute rounded-full bg-foreground"
                        style={{
                            width: particle.size,
                            height: particle.size,
                            top: 0,
                            left: '50%',
                            marginLeft: -particle.size / 2,
                            boxShadow: `0 0 ${particle.size * 2}px var(--foreground)`,
                        }}
                        animate={{
                            scale: [1, 1.5, 1],
                            opacity: [particle.opacity, particle.opacity * 1.3, particle.opacity],
                        }}
                        transition={{
                            duration: 3,
                            repeat: Infinity,
                            delay: i * 0.5,
                        }}
                    />
                </motion.div>
            ))}

            {/* The actual sphere */}
            <motion.div
                className="relative z-10"
                animate={{
                    y: [0, -12, 0],
                    rotate: [0, 2, 0, -2, 0],
                }}
                transition={{
                    y: { duration: 6, repeat: Infinity, ease: "easeInOut" },
                    rotate: { duration: 8, repeat: Infinity, ease: "easeInOut" },
                }}
            >
                <motion.div
                    initial={{ scale: 0.8, opacity: 0 }}
                    animate={{ scale: 1, opacity: 1 }}
                    transition={{ duration: 1.2, ease: [0.25, 0.1, 0, 1] }}
                >
                    <Image
                        src="/logo.png"
                        alt="SocialSphere"
                        width={300}
                        height={300}
                        className="drop-shadow-[0_0_60px_rgba(168,85,247,0.4)]"
                        priority
                    />
                </motion.div>
            </motion.div>

            {/* Sparkle effects */}
            {[175, 190, 180, 195, 170, 185, 200, 178].map((radius, i) => {
                const angle = (i / 8) * Math.PI * 2;
                return (
                    <motion.div
                        key={`sparkle-${i}`}
                        className="absolute w-1 h-1 bg-foreground rounded-full"
                        style={{
                            left: `calc(50% + ${Math.cos(angle) * radius}px)`,
                            top: `calc(50% + ${Math.sin(angle) * radius}px)`,
                        }}
                        animate={{
                            scale: [0, 1, 0],
                            opacity: [0, 1, 0],
                        }}
                        transition={{
                            duration: 2,
                            repeat: Infinity,
                            delay: i * 0.3,
                            ease: "easeInOut",
                        }}
                    />
                );
            })}
        </div>
    );
}

// Feature Card Component
function FeatureCard({ icon: Icon, title, description, className = "", index }) {
    return (
        <motion.div
            variants={featureCardVariants}
            initial="hidden"
            whileInView="visible"
            viewport={{ once: true, amount: 0.3 }}
            transition={{ delay: index * 0.1 }}
            whileHover={{ y: -4, transition: { duration: 0.2 } }}
            className={`group relative p-8 rounded-2xl border border-(--border) bg-(--muted)/5 hover:bg-(--muted)/10 hover:border-(--accent)/30 transition-all duration-300 ${className}`}
        >
            <div className="absolute inset-0 rounded-2xl bg-(--accent)/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
            <div className="relative">
                <div className="w-12 h-12 rounded-xl bg-(--accent)/10 flex items-center justify-center mb-5 group-hover:bg-(--accent)/20 transition-colors">
                    <Icon className="w-6 h-6 text-(--accent)" />
                </div>
                <h3 className="text-xl font-semibold text-foreground mb-3">{title}</h3>
                <p className="text-(--muted) leading-relaxed">{description}</p>
            </div>
        </motion.div>
    );
}

// Large Feature Card
function LargeFeatureCard({ icon: Icon, label, title, description, index }) {
    return (
        <motion.div
            variants={featureCardVariants}
            initial="hidden"
            whileInView="visible"
            viewport={{ once: true, amount: 0.3 }}
            transition={{ delay: index * 0.15 }}
            whileHover={{ y: -6, transition: { duration: 0.2 } }}
            className="group relative p-10 rounded-3xl border border-(--border) bg-(--muted)/5 hover:bg-(--muted)/10 hover:border-(--accent)/30 transition-all duration-300 overflow-hidden"
        >
            {/* Background accent glow */}
            <div className="absolute -top-24 -right-24 w-48 h-48 bg-(--accent)/10 rounded-full blur-3xl opacity-0 group-hover:opacity-100 transition-opacity duration-500" />

            <div className="relative">
                <span className="text-xs font-medium uppercase tracking-widest text-(--accent) mb-4 block">{label}</span>
                <div className="flex items-start justify-between gap-6">
                    <div className="flex-1">
                        <h3 className="text-3xl font-bold text-foreground mb-4 tracking-tight">{title}</h3>
                        <p className="text-lg text-(--muted) leading-relaxed">{description}</p>
                    </div>
                    <div className="w-16 h-16 rounded-2xl bg-(--accent)/10 flex items-center justify-center shrink-0 group-hover:bg-(--accent)/20 transition-colors">
                        <Icon className="w-8 h-8 text-(--accent)" />
                    </div>
                </div>
            </div>
        </motion.div>
    );
}

export default function LandingPage() {
    return (
        <div className="min-h-screen bg-background">
            {/* Navigation */}
            <motion.nav
                initial={{ y: -20, opacity: 0 }}
                animate={{ y: 0, opacity: 1 }}
                transition={{ duration: 0.6, ease: [0.25, 0.1, 0, 1] }}
                className="fixed top-0 inset-x-0 z-50 px-6 py-4"
            >
                <div className="max-w-6xl mx-auto">
                    <div className="flex items-center justify-between px-6 py-3 rounded-full border border-(--border) bg-(--background)/80 backdrop-blur-xl">
                        <Link href="/" className="flex items-center gap-2">
                            <Image
                                src="/logo.png"
                                alt="SocialSphere"
                                width={28}
                                height={28}
                                className="drop-shadow-[0_0_8px_rgba(168,85,247,0.3)]"
                            />
                            <span className="text-base font-semibold tracking-tight text-foreground">
                                SocialSphere
                            </span>
                        </Link>
                        <div className="hidden md:flex items-center gap-8">
                            <Link href="#features" className="text-sm text-(--muted) hover:text-foreground transition-colors">
                                Features
                            </Link>
                            <Link href="/about" className="text-sm text-(--muted) hover:text-foreground transition-colors">
                                About
                            </Link>
                        </div>
                        <div className="flex items-center gap-4">
                            <Link
                                href="/login"
                                className="text-sm text-(--muted) hover:text-foreground transition-colors"
                            >
                                Sign In
                            </Link>
                            <Link
                                href="/register"
                                className="px-5 py-2 text-sm font-medium bg-(--accent) text-white rounded-full hover:bg-(--accent-hover) transition-colors"
                            >
                                Get Started
                            </Link>
                        </div>
                    </div>
                </div>
            </motion.nav>

            {/* Hero Section */}
            <section className="relative min-h-screen flex items-center pt-24 overflow-hidden">
                {/* Background subtle grid */}
                <div className="absolute inset-0 bg-[linear-gradient(to_right,var(--border)_1px,transparent_1px),linear-gradient(to_bottom,var(--border)_1px,transparent_1px)] bg-size-[4rem_4rem] opacity-30" />

                {/* Gradient overlay */}
                <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,transparent_0%,var(--background)_70%)]" />

                <div className="relative max-w-7xl mx-auto px-6 w-full">
                    <div className="grid lg:grid-cols-2 gap-12 items-center">
                        {/* Left: Text Content */}
                        <motion.div
                            variants={containerVariants}
                            initial="hidden"
                            animate="visible"
                            className="max-w-xl"
                        >
                            <motion.div variants={itemVariants} className="mb-6">
                                <span className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-(--accent)/10 text-(--accent) text-sm font-medium">
                                    
                                    Welcome to the Sphere
                                </span>
                            </motion.div>

                            <motion.h1
                                variants={itemVariants}
                                className="text-5xl md:text-6xl lg:text-7xl font-bold tracking-tight leading-[1.1] text-foreground mb-6"
                            >
                                Your world.
                                <br />
                                <span>Your </span>
                                <span
                                    className="text-image-fill"
                                    // style={{ backgroundImage: 'url(/fill.png)' }}
                                >
                                    Sphere
                                </span>
                                <span>.</span>
                            </motion.h1>

                            <motion.p
                                variants={itemVariants}
                                className="text-xl text-(--muted) leading-relaxed mb-10"
                            >
                                A social network built for meaningful connections.
                                Share moments, join communities, and stay connected
                                with the people who matter most.
                            </motion.p>

                            <motion.div variants={itemVariants} className="flex flex-col sm:flex-row gap-4">
                                <Link
                                    href="/register"
                                    className="group inline-flex items-center justify-center gap-2 px-8 py-4 text-base font-medium bg-(--accent) text-white rounded-full hover:bg-(--accent-hover) transition-all"
                                >
                                    Join the Sphere
                                    <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
                                </Link>
                                <Link
                                    href="/login"
                                    className="inline-flex items-center justify-center px-8 py-4 text-base font-medium text-foreground rounded-full border border-(--border) hover:bg-(--muted)/10 transition-colors"
                                >
                                    Sign In
                                </Link>
                            </motion.div>
                        </motion.div>

                        {/* Right: Hero Sphere */}
                        <motion.div
                            initial={{ opacity: 0, scale: 0.8 }}
                            animate={{ opacity: 1, scale: 1 }}
                            transition={{ duration: 1, delay: 0.5, ease: [0.25, 0.1, 0, 1] }}
                            className="hidden lg:flex justify-center items-center"
                        >
                            <HeroSphere />
                        </motion.div>
                    </div>
                </div>

                {/* Scroll indicator */}
                <motion.div
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    transition={{ delay: 1.5 }}
                    className="absolute bottom-8 left-1/2 -translate-x-1/2"
                >
                    <motion.div
                        animate={{ y: [0, 8, 0] }}
                        transition={{ duration: 1.5, repeat: Infinity }}
                        className="w-6 h-10 rounded-full border-2 border-(--muted)/30 flex items-start justify-center p-2"
                    >
                        <motion.div className="w-1.5 h-1.5 bg-(--muted) rounded-full" />
                    </motion.div>
                </motion.div>
            </section>

            {/* Features Section */}
            <section id="features" className="py-32 px-6">
                <div className="max-w-6xl mx-auto">
                    {/* Section Header */}
                    <motion.div
                        initial={{ opacity: 0, y: 40 }}
                        whileInView={{ opacity: 1, y: 0 }}
                        viewport={{ once: true }}
                        transition={{ duration: 0.6 }}
                        className="text-center mb-20"
                    >
                        <span className="text-xs font-medium uppercase tracking-widest text-(--accent) mb-4 block">Features</span>
                        <h2 className="text-4xl md:text-5xl font-bold tracking-tight text-foreground mb-6">
                            Everything you need
                            <br />
                            <span className="text-(--muted)">to stay connected</span>
                        </h2>
                    </motion.div>

                    {/* Large Feature Cards */}
                    <div className="grid md:grid-cols-2 gap-6 mb-6">
                        <LargeFeatureCard
                            icon={Globe}
                            label="01 / Feeds"
                            title="Public & Friends"
                            description="Switch between public discoveries and your close circle. See what the world shares, or focus on updates from friends."
                            index={0}
                        />
                        <LargeFeatureCard
                            icon={MessageCircle}
                            label="02 / Messaging"
                            title="Real-time Chat"
                            description="Instant messaging that feels natural. Private conversations, group chats, and seamless notifications keep you connected."
                            index={1}
                        />
                    </div>

                    {/* Bento Grid */}
                    <div className="grid md:grid-cols-3 gap-6">
                        <FeatureCard
                            icon={Users}
                            title="Groups"
                            description="Create communities around shared interests. Plan events, share posts, and grow together."
                            index={0}
                        />
                        <FeatureCard
                            icon={Bell}
                            title="Notifications"
                            description="Real-time alerts keep you in the loop. Never miss a thing."
                            index={1}
                        />
                        <FeatureCard
                            icon={Shield}
                            title="Privacy First"
                            description="Control who sees your content. Public, friends-only, or completely private—your choice."
                            index={2}
                        />
                    </div>
                </div>
            </section>


            {/* CTA Section */}
            <section className="pt-12 pb-32 px-6">
                <div className="max-w-4xl mx-auto text-center">
                    <motion.div
                        initial={{ opacity: 0, y: 40 }}
                        whileInView={{ opacity: 1, y: 0 }}
                        viewport={{ once: true }}
                        transition={{ duration: 0.6 }}
                    >
                        <div className="relative inline-flex items-center justify-center mb-8">
                            <Image
                                src="/logo.png"
                                alt=""
                                width={80}
                                height={80}
                                className="drop-shadow-[0_0_30px_rgba(168,85,247,0.5)]"
                            />
                        </div>
                        <h2 className="text-4xl md:text-5xl font-bold tracking-tight text-foreground mb-6">
                            Ready to join
                            <br />
                            <span>the </span>
                            <span
                                className="text-image-fill"
                            >
                                sphere
                            </span>
                            <span>?</span>
                        </h2>
                        <p className="text-xl text-(--muted) mb-10 max-w-2xl mx-auto">
                            Connect with friends, discover communities, and share what matters.
                        </p>
                        <Link
                            href="/register"
                            className="group inline-flex items-center gap-2 px-10 py-4 text-lg font-medium bg-(--accent) text-white rounded-full hover:bg-(--accent-hover) transition-all"
                        >
                            Create Your Account
                            <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
                        </Link>
                    </motion.div>
                </div>
            </section>

            
        </div>
    );
}
