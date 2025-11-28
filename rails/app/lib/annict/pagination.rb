# typed: false
# frozen_string_literal: true

module Annict
  class Pagination
    attr_reader :after, :before

    def initialize(before:, after:, per:)
      @before = before
      @after = after
      @per = per
    end

    def first
      return per if after
      return per if !before && !after

      nil
    end

    def last
      before ? per : nil
    end

    private

    attr_reader :per
  end
end
