import speech_recognition as sr

recognizer = sr.Recognizer()

with sr.Microphone() as source:
    print("말해주세요...")
    audio = recognizer.listen(source)

try:
    print("음성을 인식한 텍스트: " + recognizer.recognize_google(audio, language='ko-KR'))
except sr.UnknownValueError:
    print("음성을 인식할 수 없습니다.")
except sr.RequestError as e:
    print(f"Google 음성 인식 서비스에 접근할 수 없습니다: {e}")
