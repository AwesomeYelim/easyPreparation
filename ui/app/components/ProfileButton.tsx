export default function ProfileButton({
  image,
  onClick,
}: {
  image: string;
  onClick: () => void;
}) {
  return (
    <div style={{ position: "relative", display: "inline-block" }}>
      <img
        src={image}
        alt="profile-img"
        style={{
          width: 25,
          height: 25,
          borderRadius: "50%",
          cursor: "pointer",
        }}
        onClick={onClick}
      />
    </div>
  );
}
