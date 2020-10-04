# frozen_string_literal: true

module Forms
  class EpisodeRecordFormComponent < ApplicationComponent
    def initialize(form:, viewer:)
      @form = form
      @viewer = viewer
    end

    private

    def form_method
      @form.record_id ? :patch : :post
    end
  end
end
