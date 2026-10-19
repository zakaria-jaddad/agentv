NAME1		= agentv-1
NAME2		= agentv-2
SRC			= ./cmd/agentv

.PHONY: all clean fclean re

all: $(NAME1) $(NAME2)

$(NAME1):
	go build -ldflags="-s -w" -o $@ $(SRC)

$(NAME2):
	go build -ldflags="-s -w" -o $@ $(SRC)

clean:
	$(RM) *.out *.coverprofile profile.cov *.test

fclean: clean
	$(RM) $(NAME1) $(NAME2)

re: fclean all