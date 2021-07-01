# frozen_string_literal: true

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
