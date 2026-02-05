"use client"

import { motion } from "motion/react"

export default function LoadingThreeDotsJumping() {
    const dotVariants = {
        jump: {
            y: -10,
            transition: {
                duration: 0.5,
                repeat: Infinity,
                repeatType: "reverse",
                ease: "easeInOut",
            },
        },
    }

    return (
        <motion.div
            animate="jump"
            transition={{ staggerChildren: -0.2, staggerDirection: -1 }}
            className="container"
        >
            <motion.div className="dot" variants={dotVariants} />
            <motion.div className="dot" variants={dotVariants} />
            <motion.div className="dot" variants={dotVariants} />
            <StyleSheet />
        </motion.div>
    )
}

/**
 * ==============   Styles   ================
 */
function StyleSheet() {
    return (
        <style>
            {`
            .container {
                display: flex;
                justify-content: center;
                align-items: center;
                gap: 4px;
                height: 1.5em;
                padding: 0.25em 0;
            }

            .dot {
                width: 8px;
                height: 8px;
                border-radius: 50%;
                background-color: currentColor;
                will-change: transform;
            }
            `}
        </style>
    )
}
