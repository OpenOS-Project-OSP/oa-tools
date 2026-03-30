# oa: un motore in C per il remastering

Il progetto è nato dalla scoperta di MX-arch di AdrianMX segnalatami da un sul mio gruppo telegram: https://t.me/penguins_eggs da cui è scaturita una proposta ad AdrianTX stesso https://github.com/AdrianTM/mx-snapshot/issues/20 e che mi ha spiegato che il suo tool MX-snapshot è il vero antenato sia di penguins-eggs che di refracta-snapshot

## Filosofia
- usiamo OverlayFS per proiettare il filestem reale nella liveroot e renderla scrivibile (come in [penguins-eggs](https://github.com/pieroproietti/penguins-eggs));
- contiene solo lo stretto necessario che può essere messo a fattor comune per la rimasterizzazione di - idealmente - tutte le distro (arch, debian, manjaro, rhel, opensuse);
- Usare la filosofia yocto per la creazione utenti passando per quanto già fatto in [penguins-eggs](https://github.com/pieroproietti/penguins-eggs/blob/master/src/classes/users.ts)

# Stato
- abbiamo la ISO che si avvia

# Prossimo obiettivo
- Implementare la action_users con il DNA di Yocto per gestire gli utenti nel chroot.

  