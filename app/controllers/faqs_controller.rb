# frozen_string_literal: true

class FaqsController < ApplicationV6Controller
  def show
    redirect_to "https://github.com/kiraka/annict/blob/main/docs/faq.md"
  end
end
