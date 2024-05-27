# typed: false
# frozen_string_literal: true

class FaqsController < ApplicationV6Controller
  def show
    redirect_to "https://github.com/annict/annict/blob/main/docs/faq.md"
  end
end
