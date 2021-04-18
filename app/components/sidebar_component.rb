# frozen_string_literal: true

class SidebarComponent < ApplicationComponent
  def initialize(search:)
    @search = search
  end
end
