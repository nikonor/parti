package BG

import l "parti/logger"

func loggerChans(errCh chan string, debugChan chan string, infoChan chan string, doneChan <-chan struct{}) {
	for {
		select {
		case s := <-errCh:
			l.Err(s)
		case s := <-debugChan:
			l.Debug(s)
		case s := <-infoChan:
			l.Info(s)
		case <-doneChan:
			l.Debug("Завершаем errorChan")
			return
		}
	}
}
