
rm ./main -f
rm ./main.I* -f
rm ./main.W* -f
rm ./main.E* -f
rm ./nohup.out -f

go build ./main.go

nohup ./main --log_dir=./ &
