"use client";

import { useState } from "react";
import Image from "next/image";

export default function PostImage({ src, alt = "Post content" }) {
    const [isLoading, setIsLoading] = useState(true);

    if (!src) return null;

    return (
        <div className="w-full mt-2 flex justify-center">
            <div className="relative max-w-full rounded-xl overflow-hidden bg-black/5 group/image">
                {/* Blurred Background Layer - Fills the space of the image */}
                <div
                    className="absolute inset-0 bg-cover bg-center blur-2xl opacity-50 scale-110 transition-opacity duration-700"
                    style={{ backgroundImage: `url(${src})` }}
                    aria-hidden="true"
                />

                {/* Main Image */}
                <Image
                    src={src}
                    alt={alt}
                    width={0}
                    height={0}
                    sizes="100vw"
                    unoptimized
                    style={{
                        width: 'auto',
                        height: 'auto',
                        maxWidth: '100%',
                        maxHeight: '600px',
                        display: 'block'
                    }}
                    className={`
                        object-contain relative z-10 
                        transition-all duration-500 ease-out
                        ${isLoading ? "opacity-0 scale-95" : "opacity-100 scale-100"}
                    `}
                    onLoad={() => setIsLoading(false)}
                />
            </div>
        </div>
    );
}
