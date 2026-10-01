// Cronômetro sincronizado com o servidor (API-02; spec 06, CA-05 a CA-07; plano, D5).
// Todos os instantes em ms desde a época.

// calcularDefasagem devolve quanto o relógio do servidor está à frente do local.
export function calcularDefasagem(horarioServidor: string, recebidaEmLocal: number): number {
  return Date.parse(horarioServidor) - recebidaEmLocal;
}

// tempoRestante devolve os ms até o fim da fase no relógio do servidor, nunca negativo,
// ou null quando não há fim de fase (ENCERRADA).
export function tempoRestante(fimDaFase: string | undefined, defasagem: number, agoraLocal: number): number | null {
  if (fimDaFase === undefined) return null;
  return Math.max(0, Date.parse(fimDaFase) - (agoraLocal + defasagem));
}
