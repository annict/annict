# frozen_string_literal: true

module V6
  class FaqsController < V6::ApplicationController
    def show
      redirect_to "https://github.com/kiraka/annict/blob/main/docs/faq.md"
    end
  end
end
