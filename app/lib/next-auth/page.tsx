import { useEffect, useState } from "react";
import { useSession } from "next-auth/react";
import { UserChurchInfo } from "@/components/SideBar";
import { useRecoilState } from "recoil";
import { userInfoState } from "@/recoilState";

export default function CompleteProfile() {
  const { data: session, status } = useSession();
  const [nameInfo, setNameInfo] = useState({ name: "", englishName: "" });
  const [userInfo, setUserInfo] = useRecoilState(userInfoState);
  const [close, setClose] = useState(true);

  useEffect(() => {
    const checkUser = async () => {
      if (!session?.user?.email) return;

      const response = await fetch(`/api/user?email=${session.user.email}`);
      if (!response.ok) return;

      const data: UserChurchInfo = await response.json();
      setUserInfo(data);
      setClose(false);
      setNameInfo({ name: data.name, englishName: data.english_name });
    };

    if (session?.user?.email) {
      checkUser();
    }
  }, [session]);

  const handleSubmit = async () => {
    await fetch("/api/user", {
      method: "POST",
      body: JSON.stringify({
        name: nameInfo.name,
        english_name: nameInfo.englishName,
        email: session?.user?.email,
      }),
      headers: { "Content-Type": "application/json" },
    });
    setClose(true);
    setUserInfo((prev) => ({ ...prev, ...nameInfo }));
  };

  if (status === "loading" || close) return null;

  if (!userInfo?.english_name || !userInfo?.name) {
    return (
      <div
        style={{
          position: "fixed",
          top: 0,
          left: 0,
          width: "100vw",
          height: "100vh",
          backgroundColor: "rgba(0, 0, 0, 0.4)",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          zIndex: 1000,
        }}
      >
        <div
          style={{
            maxWidth: "400px",
            width: "90%",
            padding: "20px",
            border: "1px solid #ccc",
            borderRadius: "8px",
            backgroundColor: "#fff",
            display: "flex",
            flexDirection: "column",
            gap: "10px",
            boxShadow: "0 4px 12px rgba(0,0,0,0.15)",
          }}
        >
          <h2 style={{ textAlign: "center", marginBottom: "10px" }}>
            추가 정보 입력
          </h2>
          <input
            placeholder="교회 이름"
            value={nameInfo.name}
            onChange={(e) => setNameInfo({ ...nameInfo, name: e.target.value })}
            style={{
              padding: "8px",
              borderRadius: "4px",
              border: "1px solid #ccc",
            }}
          />
          <input
            placeholder="영문 이름"
            value={nameInfo.englishName}
            onChange={(e) =>
              setNameInfo({ ...nameInfo, englishName: e.target.value })
            }
            style={{
              padding: "8px",
              borderRadius: "4px",
              border: "1px solid #ccc",
            }}
          />
          <button
            onClick={handleSubmit}
            style={{
              padding: "10px",
              backgroundColor: "#0070f3",
              color: "#fff",
              border: "none",
              borderRadius: "4px",
              fontWeight: "bold",
              cursor: "pointer",
            }}
          >
            저장
          </button>
        </div>
      </div>
    );
  }

  return null;
}
