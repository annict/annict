# frozen_string_literal: true

class WorkSubNavComponent < ApplicationComponent
  def initialize(work:, page_category:)
    @work = work
    @page_category = page_category
  end

  def call
    Htmlrb.build do |el|
      el.div class: "c-subnav c-subnav--transparent" do
        el.a(
          href: "/works/#{work.id}",
          class: "c-subnav__link #{'c-subnav__link--active' if page_category == 'work_detail'}"
        ) do
          el.div class: "c-subnav__item" do
            t "noun.detail"
          end
        end

        unless work.is_no_episodes
          el.a(
            href: "/works/#{work.id}/episodes",
            class: "c-subnav__link #{'c-subnav__link--active' if page_category == 'episode_list'}"
          ) do
            el.div class: "c-subnav__item" do
              t "noun.episodes"
            end
          end
        end

        el.a(
          href: "/works/#{work.id}/records",
          class: "c-subnav__link #{'c-subnav__link--active' if page_category == 'work_record_list'}"
        ) do
          el.div class: "c-subnav__item" do
            t "noun.records"
          end
        end
      end
    end.html_safe
  end

  private

  attr_reader :work, :page_category
end
