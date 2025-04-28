"use client";

import { useEffect, useState } from "react";

type Post = {
  id: number;
  title: string;
  content: string;
};

export default function PostList() {
  const [posts, setPosts] = useState<Post[]>([]);

  useEffect(() => {
    async function fetchPosts() {
      const response = await fetch("/api/posts");
      const data: Post[] = await response.json();
      setPosts(data);
    }

    fetchPosts();
  }, []);

  return (
    <div>
      {/* <h1>Posts</h1> */}
      <ul>
        {posts.map((post) => (
          <li key={post.id}>
            <h2>{post.title}</h2>
            <p>{post.content}</p>
          </li>
        ))}
      </ul>
    </div>
  );
}
