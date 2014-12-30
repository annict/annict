json.partial! '/works/episodes', episodes: @work.episodes.order(:sort_number), work: @work, user: current_user
