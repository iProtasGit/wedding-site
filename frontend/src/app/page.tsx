"use client";

import { useEffect, useState } from "react";

type Config = {
  names: string;
  weddingDate: string;
  dateText: string;
  timeText: string;
  venue: string;
  transferInfo: string;
  costumeInfo: string;
  apiEndpoint: string;
};

type GuestForm = {
  id: number;
  fullName: string;
  alcohol: string[];
  otherAlcohol: string;
  transfer: boolean;
};

const alcoholOptions = [
  "Вино",
  "Водка",
  "Шампанское",
  "Виски",
  "Другое",
  "Не пью",
];

export default function Home() {
  const [config, setConfig] = useState<Config | null>(null);
  const [timeLeft, setTimeLeft] = useState({
    days: 0,
    hours: 0,
    minutes: 0,
    seconds: 0,
  });
  const [guests, setGuests] = useState<GuestForm[]>([
    { id: Date.now(), fullName: "", alcohol: [], otherAlcohol: "", transfer: false },
  ]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [message, setMessage] = useState<{ text: string; type: "success" | "error" | "warning" } | null>(null);

  useEffect(() => {
    fetch("/config.json")
      .then((res) => res.json())
      .then((data) => {
        setConfig(data);
      })
      .catch((err) => console.error("Failed to load config", err));
  }, []);

  useEffect(() => {
    if (!config?.weddingDate) return;

    const target = new Date(config.weddingDate).getTime();
    const interval = setInterval(() => {
      const now = new Date().getTime();
      const distance = target - now;

      if (distance < 0) {
        clearInterval(interval);
        return;
      }

      setTimeLeft({
        days: Math.floor(distance / (1000 * 60 * 60 * 24)),
        hours: Math.floor((distance % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60)),
        minutes: Math.floor((distance % (1000 * 60 * 60)) / (1000 * 60)),
        seconds: Math.floor((distance % (1000 * 60)) / 1000),
      });
    }, 1000);

    return () => clearInterval(interval);
  }, [config?.weddingDate]);

  const addGuest = () => {
    setGuests((prev) => [
      ...prev,
      { id: Date.now(), fullName: "", alcohol: [], otherAlcohol: "", transfer: false },
    ]);
  };

  const removeGuest = (id: number) => {
    setGuests((prev) => prev.filter((g) => g.id !== id));
  };

  const updateGuest = (id: number, field: keyof GuestForm, value: string | string[] | boolean) => {
    setMessage(null); // Clear errors when typing
    setGuests((prev) =>
      prev.map((g) => (g.id === id ? { ...g, [field]: value } : g))
    );
  };

  const toggleAlcohol = (id: number, option: string) => {
    setMessage(null);
    setGuests((prev) =>
      prev.map((g) => {
        if (g.id === id) {
          const hasOption = g.alcohol.includes(option);
          // If "Не пью" is selected, clear other options
          if (option === "Не пью") {
            return { ...g, alcohol: hasOption ? [] : ["Не пью"], otherAlcohol: "" };
          }
          // If selecting another option, ensure "Не пью" is removed
          let newAlcohol = hasOption
            ? g.alcohol.filter((a) => a !== option)
            : [...g.alcohol, option];
          newAlcohol = newAlcohol.filter((a) => a !== "Не пью");
          
          return { ...g, alcohol: newAlcohol };
        }
        return g;
      })
    );
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setMessage(null);

    // Client-side validation
    for (let i = 0; i < guests.length; i++) {
      const g = guests[i];
      if (g.fullName.trim() === "") {
        setMessage({ text: `Пожалуйста, введите имя для Гостя №${i + 1}.`, type: "warning" });
        setIsSubmitting(false);
        return;
      }
      if (g.alcohol.includes("Другое") && g.otherAlcohol.trim() === "") {
        setMessage({ text: `Уточните алкоголь для "${g.fullName}" (выбрано "Другое").`, type: "warning" });
        setIsSubmitting(false);
        return;
      }
    }

    const payload = {
      guests: guests.map((g) => ({
        fullName: g.fullName,
        alcohol: g.alcohol,
        otherAlcohol: g.alcohol.includes("Другое") ? g.otherAlcohol : "",
        transfer: g.transfer,
      })),
    };

    try {
      const res = await fetch(config?.apiEndpoint || "/api/rsvp", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      const data = await res.json().catch(() => ({})); // Handle non-JSON responses safely

      if (!res.ok) {
        // Display specific error message from the backend if available
        throw new Error(data.error || "Ошибка сохранения");
      }
      
      setMessage({ text: "Спасибо! Ваш ответ успешно сохранен.", type: "success" });
      setGuests([{ id: Date.now(), fullName: "", alcohol: [], otherAlcohol: "", transfer: false }]);
    } catch (err: any) {
      console.error(err);
      setMessage({ 
        text: err.message || "Сервер временно недоступен. Попробуйте еще раз позже.", 
        type: "error" 
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!config) return <div className="min-h-screen flex items-center justify-center text-pastel-text">Загрузка...</div>;

  return (
    <main className="min-h-screen flex flex-col font-inter">
      {/* Hero Section */}
      <header 
        className="relative w-full min-h-screen flex items-center justify-center text-center px-4 fade-in"
        style={{
          backgroundImage: "url('https://images.unsplash.com/photo-1520854221256-17451cc331bf?auto=format&fit=crop&q=80&w=2000')",
          backgroundSize: 'cover',
          backgroundPosition: 'center',
        }}
      >
        <div className="absolute inset-0 bg-white/40"></div>
        <div className="relative glass-panel p-8 md:p-16 rounded-2xl shadow-xl max-w-2xl w-full border border-white/50">
          <p className="uppercase tracking-widest text-sm text-pastel-text mb-4">Приглашаем на нашу свадьбу</p>
          <h1 className="text-5xl md:text-7xl mb-6 text-pastel-text font-playfair">{config.names}</h1>
          <p className="text-xl md:text-2xl mb-8 text-pastel-text/80 italic font-playfair">{config.dateText}</p>
          
          <div className="flex justify-center space-x-4 md:space-x-8 mt-8 text-pastel-text">
            <div className="flex flex-col"><span className="text-4xl md:text-5xl font-playfair font-semibold">{String(timeLeft.days).padStart(2, '0')}</span><span className="text-[10px] md:text-xs uppercase tracking-wider mt-2">Дней</span></div>
            <div className="flex flex-col"><span className="text-4xl md:text-5xl font-playfair font-semibold">{String(timeLeft.hours).padStart(2, '0')}</span><span className="text-[10px] md:text-xs uppercase tracking-wider mt-2">Часов</span></div>
            <div className="flex flex-col"><span className="text-4xl md:text-5xl font-playfair font-semibold">{String(timeLeft.minutes).padStart(2, '0')}</span><span className="text-[10px] md:text-xs uppercase tracking-wider mt-2">Минут</span></div>
            <div className="flex flex-col"><span className="text-4xl md:text-5xl font-playfair font-semibold">{String(timeLeft.seconds).padStart(2, '0')}</span><span className="text-[10px] md:text-xs uppercase tracking-wider mt-2">Секунд</span></div>
          </div>
        </div>
      </header>

      {/* Details Section */}
      <section className="py-24 px-4 bg-white text-center">
        <div className="max-w-5xl mx-auto">
          <h2 className="text-4xl mb-16 text-pastel-text font-playfair">Детали мероприятия</h2>
          <div className="grid md:grid-cols-3 gap-12 text-pastel-text">
            <div className="p-6 bg-pastel-background rounded-xl shadow-sm border border-pastel-green/20">
              <h3 className="text-2xl mb-4 font-playfair">Где и Когда</h3>
              <p className="text-lg mb-2">{config.dateText}</p>
              <p className="text-lg mb-2">Начало: {config.timeText}</p>
              <p className="text-lg mt-4 font-medium">{config.venue}</p>
            </div>
            <div className="p-6 bg-pastel-background rounded-xl shadow-sm border border-pastel-green/20">
              <h3 className="text-2xl mb-4 font-playfair">Трансфер</h3>
              <p className="text-base leading-relaxed">{config.transferInfo}</p>
            </div>
            <div className="p-6 bg-pastel-background rounded-xl shadow-sm border border-pastel-green/20">
              <h3 className="text-2xl mb-4 font-playfair">Дресс-код</h3>
              <p className="text-base leading-relaxed">{config.costumeInfo}</p>
            </div>
          </div>
        </div>
      </section>

      {/* RSVP Section */}
      <section className="py-24 px-4 bg-pastel-background">
        <div className="max-w-3xl mx-auto bg-white p-8 md:p-14 shadow-lg border border-pastel-green/30 rounded-2xl">
          <h2 className="text-4xl text-center mb-4 text-pastel-text font-playfair">Подтверждение присутствия</h2>
          <p className="text-center text-pastel-text/70 mb-10">Пожалуйста, заполните форму, чтобы мы знали, что вы будете с нами.</p>
          
          <form onSubmit={handleSubmit} className="space-y-8">
            <div className="space-y-8">
              {guests.map((guest, index) => (
                <div key={guest.id} className="relative border border-pastel-green p-6 rounded-xl bg-white shadow-sm transition-all hover:shadow-md">
                  {guests.length > 1 && (
                    <button 
                      type="button" 
                      className="absolute top-4 right-4 text-pastel-text/50 hover:text-red-400 transition-colors text-xl"
                      onClick={() => removeGuest(guest.id)}
                    >
                      &times;
                    </button>
                  )}
                  
                  <div className="mb-6">
                    <label className="block text-lg font-medium text-pastel-text mb-2">
                      {index === 0 ? "Ваше Имя и Фамилия" : `Имя и Фамилия Гостя ${index + 1}`}
                    </label>
                    <input 
                      type="text" 
                      required 
                      value={guest.fullName}
                      onChange={(e) => updateGuest(guest.id, "fullName", e.target.value)}
                      className="w-full px-4 py-3 border border-pastel-green/50 rounded-lg focus:outline-none focus:ring-2 focus:ring-pastel-greenDark focus:border-transparent text-pastel-text bg-[#fcfdfc]" 
                      placeholder="Например, Анна Петрова"
                    />
                  </div>

                  <div className="mb-6">
                    <label className="block text-base font-medium text-pastel-text mb-3">Предпочтения по напиткам</label>
                    <div className="grid grid-cols-2 sm:grid-cols-3 gap-3 text-sm text-pastel-text/80">
                      {alcoholOptions.map((option) => (
                        <label key={option} className={`flex items-center space-x-3 p-3 rounded-lg border cursor-pointer transition-colors ${guest.alcohol.includes(option) ? 'bg-pastel-green/30 border-pastel-green' : 'border-gray-100 hover:bg-gray-50'}`}>
                          <input 
                            type="checkbox" 
                            className="rounded border-pastel-green text-pastel-greenDark focus:ring-pastel-greenDark w-4 h-4 accent-pastel-greenDark"
                            checked={guest.alcohol.includes(option)}
                            onChange={() => toggleAlcohol(guest.id, option)}
                          />
                          <span>{option}</span>
                        </label>
                      ))}
                    </div>
                    
                    {guest.alcohol.includes("Другое") && (
                      <div className="mt-4 fade-in">
                        <label className="block text-sm text-pastel-text/80 mb-2">Уточните ваши предпочтения</label>
                        <input 
                          type="text" 
                          value={guest.otherAlcohol}
                          onChange={(e) => updateGuest(guest.id, "otherAlcohol", e.target.value)}
                          className="w-full px-4 py-2 text-sm border border-pastel-green/50 rounded-lg focus:outline-none focus:ring-2 focus:ring-pastel-greenDark focus:border-transparent text-pastel-text bg-[#fcfdfc]" 
                          placeholder="Напишите, что бы вы хотели..."
                          required={guest.alcohol.includes("Другое")}
                        />
                      </div>
                    )}
                  </div>

                  <div className="pt-4 border-t border-pastel-green/30">
                    <label className="flex items-center space-x-3 cursor-pointer p-2 rounded-lg hover:bg-pastel-green/10 transition-colors inline-flex">
                      <input 
                        type="checkbox" 
                        checked={guest.transfer}
                        onChange={(e) => updateGuest(guest.id, "transfer", e.target.checked)}
                        className="rounded border-pastel-green text-pastel-greenDark focus:ring-pastel-greenDark w-5 h-5 accent-pastel-greenDark"
                      />
                      <span className="text-pastel-text font-medium text-lg">Нужен трансфер?</span>
                    </label>
                  </div>
                </div>
              ))}
            </div>

            <div className="flex justify-center pt-2">
              <button 
                type="button" 
                onClick={addGuest}
                className="py-2.5 px-6 rounded-lg border-2 border-pastel-greenDark text-pastel-greenDark hover:bg-pastel-greenDark hover:text-white transition-all text-sm font-medium"
              >
                + Добавить гостя
              </button>
            </div>

            {message && (
              <div className={`mt-6 text-center py-4 px-6 rounded-xl font-medium ${
                message.type === 'success' ? 'bg-green-100 text-green-800 border border-green-200' : 
                message.type === 'warning' ? 'bg-yellow-50 text-yellow-800 border border-yellow-200' :
                'bg-red-50 text-red-800 border border-red-200'
              }`}>
                {message.text}
              </div>
            )}

            <button 
              type="submit" 
              disabled={isSubmitting}
              className="w-full py-4 rounded-xl text-lg font-playfair tracking-wide mt-6 bg-pastel-greenDark text-white hover:bg-[#8b9e8a] transition-all disabled:opacity-70 shadow-md hover:shadow-lg"
            >
              {isSubmitting ? "Отправка..." : "Отправить ответ"}
            </button>
          </form>
        </div>
      </section>

      <footer className="py-12 bg-white text-center text-pastel-text/60 text-sm border-t border-pastel-green/20">
        <p className="font-playfair italic text-3xl mb-3 text-pastel-text">{config.names}</p>
        <p className="text-base">С нетерпением ждем встречи с вами!</p>
      </footer>
    </main>
  );
}
