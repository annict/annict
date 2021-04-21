# frozen_string_literal: true

module Forms
  class EpisodeRecordFormComponent < ApplicationComponent
    def initialize(form:, current_user:)
      @form = form
      @current_user = current_user
    end

    private

    def form_method
      @form.persisted? ? :patch : :post
    end

    def form_url
      @form.persisted? ? my_record_path(@current_user.username, @form.record.id) : my_episode_record_list_path(@form.episode.id)
    end
  end
end
