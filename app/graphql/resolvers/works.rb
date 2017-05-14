# frozen_string_literal: true

module Resolvers
  class Works
    def call(_obj, args, _ctx)
      @collection = Work.all
      from_arguments(args)
    end

    private

    def from_arguments(args = {})
      %i(
        annictIds
        titles
        seasons
      ).each do |arg_name|
        next if args[arg_name].blank?
        @collection = send(arg_name.to_s.underscore, args[arg_name])
      end

      @collection
    end

    def annict_ids(ids)
      @collection.where(id: ids)
    end

    def titles(titles)
      @collection.search(title_or_title_kana_cont_any: titles).result
    end

    def seasons(seasons)
      @collection.by_seasons(seasons)
    end
  end
end
