# frozen_string_literal: true

module Selectors
  class ProgramSelectorComponent < ApplicationComponent
    def initialize(anime:, library_entry:, class_name: "")
      @anime = anime
      @library_entry = library_entry
      @class_name = class_name
      @user = library_entry.user
    end

    private

    def program_selector_class_name
      classes = []
      classes += @class_name.split(" ")
      classes.uniq.join(" ")
    end

    def program_options
      options = @anime.
        programs.
        only_kept.
        eager_load(:channel).
        merge(@user.channels.only_kept).
        order(:started_at).
        map do |program|
          name = [program.channel.name]
          name << "#{helpers.display_time(program.started_at)}~" if program.started_at.present?
          [name.join(" | "), program.id]
        end

        options.insert(0, [t("messages._components.program_selector.select_program"), "no_select"])
    end
  end
end
