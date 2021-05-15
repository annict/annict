# frozen_string_literal: true

class FaqsController < ApplicationController
  def index
    @faq_categories = FaqCategory.only_kept.order(:sort_number)
  end
end
