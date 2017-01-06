# frozen_string_literal: true

class MutesController < ApplicationController
  before_action :authenticate_user!

  def index
    @mute_users = current_user.mute_users.order(id: :desc)
  end
end
