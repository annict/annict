# typed: false
# frozen_string_literal: true

class CommunitiesController < ApplicationV6Controller
  def show
    redirect_to(ENV.fetch("ANNICT_COMMUNITY_URL"), allow_other_host: true)
  end
end
