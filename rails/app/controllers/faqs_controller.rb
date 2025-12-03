# typed: false
# frozen_string_literal: true

class FaqsController < ApplicationV6Controller
  def show
    redirect_to "https://github.com/annict/annict/blob/main/rails/docs/faq.md", allow_other_host: true
  end
end
