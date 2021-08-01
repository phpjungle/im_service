BASEDIR=$HOME/go/src/github.com/GoBelieveIO/im_service
CFGPATH=$BASEDIR/scripts/phpjungle/
IMPATH=$HOME/im

nohup $BASEDIR/bin/ims -log_dir=$IMPATH/logs/ims $CFGPATH/ims.cfg >$IMPATH/logs/ims/ims.log 2>&1 &

nohup $BASEDIR/bin/imr -log_dir=$IMPATH/logs/imr $CFGPATH/imr.cfg >$IMPATH/logs/imr/imr.log 2>&1 &

nohup $BASEDIR/bin/im -log_dir=$IMPATH/logs/im $CFGPATH/im.cfg >$IMPATH/logs/im/im.log 2>&1 &