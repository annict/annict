# frozen_string_literal: true

module DB
  class EpisodeFormCollection < Array
    def <<(episode)
      if episode.is_a?(Hash)
        super(DB::EpisodeForm.new(episode))
      else
        super
      end
    end
  end

  class EpisodesForm
    include ActiveModel::Model
    include Virtus.model
    include DbActivityMethods

    attribute :episodes, DB::EpisodeFormCollection[DB::EpisodeForm]

    def self.load(work)
      @work = work
      @episodes = work.
        episodes.
        published.
        order(sort_number: :desc).
        map { |e| DB::EpisodeForm.new(e) }

      new(episodes: @episodes)
    end
  end
end
