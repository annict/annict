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

    attribute :episode_forms, DB::EpisodeFormCollection[DB::EpisodeForm]

    def self.load(work, params_list = [])
      @work = work

      episodes = work.episodes.published.order(sort_number: :desc)
      @episode_forms = (params_list.presence || episodes).map do |obj|
        form = DB::EpisodeForm.new(obj)
        form.root_resource = @work
        form
      end

      new episode_forms: @episode_forms
    end

    def valid?
      @episode_forms.all?(&:valid?)
    end

    def errors
      @episode_forms.map(&:errors).select(&:present?).first.presence || []
    end

    def save_and_create_activity!(user)
      ActiveRecord::Base.transaction do
        episode_forms.each do |form|
          form.save_and_create_activity!(user)
        end
      end
    end
  end
end
