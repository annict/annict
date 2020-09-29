# frozen_string_literal: true

module Forms
  class EpisodeRecordFormComponent < ApplicationComponent
    def initialize(form:)
      @form = form
    end

    private

    attr_reader :form

    def form_method
      form.record_id ? :patch : :post
    end
  end
end
