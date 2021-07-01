# frozen_string_literal: true

class SidebarComponent < ApplicationComponent
  def initialize(user:, search:)
    @user = user
    @search = search
  end

  private

  attr_reader :user, :search
end
