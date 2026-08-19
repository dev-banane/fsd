#ifndef WIN32
	#include <unistd.h>
#endif
#include <cstdio>
#include <cstdlib>
#include <cstring>

#include "process.h"
#include "global.h"
#include "server.h"
#include "support.h"

int process::calcmasks(fd_set *, fd_set *) { return 0; }
int process::run() { return 0; }
process::process()
{
   next=prev=NULL;
}


pman::pman()
{
   rootprocess=NULL;
   busy=0;
}
void pman::registerprocess(process *what)
{
   what->next=rootprocess;
   if (what->next) what->next->prev=what;
   rootprocess=what;
}
void pman::run()
{
   int max=0;
   process *temp;
   struct timeval timeout;
   fd_set rmask, wmask;
   FD_ZERO(&rmask); FD_ZERO(&wmask);
   for (temp=rootprocess;temp;temp=temp->next)
   {
      int num=temp->calcmasks(&rmask, &wmask);
      if (num>max) max=num;
   }
   timeout.tv_usec=0, timeout.tv_sec=(busy?0:1);
#ifndef WIN32
   struct timeval tracestart;
   if (traceenabled) gettimeofday(&tracestart, NULL);
#endif
   int selresult=select(max+1, &rmask, &wmask, NULL, &timeout);
   if (selresult<=0)
   {
      FD_ZERO(&rmask); FD_ZERO(&wmask);
   }
   busy=0;
   for (temp=rootprocess;temp;temp=temp->next)
   {
      temp->rmask=&rmask, temp->wmask=&wmask;
      if (temp->run()) busy=1;
   }
#ifndef WIN32
   if (traceenabled)
   {
      struct timeval traceend;
      gettimeofday(&traceend, NULL);
      long elapsedms=(traceend.tv_sec-tracestart.tv_sec)*1000+
         (traceend.tv_usec-tracestart.tv_usec)/1000;
      if (selresult>0||elapsedms>50)
         tracelog("LOOP select=%d elapsedms=%ld busy=%d", selresult,
            elapsedms, busy);
   }
#endif
}
