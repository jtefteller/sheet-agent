DIR=.sheets-agent

install:
	@mkdir -p $(HOME)/$(DIR)
	@echo "Please install google auth credentials in $(HOME)/$(DIR)/credentials.json"

refresh:
	@rm $(HOME)/$(DIR)/token.json
