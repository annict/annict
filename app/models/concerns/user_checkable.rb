module UserCheckable
  extend ActiveSupport::Concern

  included do
    def checkedin?(episode)
      checkins.exists?(episode_id: episode.id)
    end

    def checkins_count(episode)
      checkins.where(episode: episode).count
    end
  end
end
