# frozen_string_literal: true

class ApplicationRepository
  def initialize(viewer:)
    @viewer = viewer
  end

  def load_query(path)
    File.read(Rails.root.join("app/queries/#{path}"))
  end

  private

  attr_reader :viewer
end
