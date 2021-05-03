# frozen_string_literal: true

module Forms
  class AnimeRecordFormComponent < ApplicationComponent
    def initialize(form:, viewer:)
      @form = form
      @viewer = viewer
    end

    private

    def form_method
      @form.persisted? ? :patch : :post
    end
  end
end
