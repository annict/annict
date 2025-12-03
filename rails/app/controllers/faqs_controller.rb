# typed: false
# frozen_string_literal: true

class FaqsController < ApplicationV6Controller
  def show
    redirect_to "https://wikino.app/s/annict/pages/323", allow_other_host: true
  end
end
