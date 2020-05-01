# frozen_string_literal: true

class TabBarComponent < ApplicationComponent
  def initialize(user:)
    @user = user
  end

  def call
    Htmlrb.build do |el|
      el.c_tab_bar class: "bg-white", v_slot: "tabBar" do
        el.div class: "d-block h-100 navbar navbar-expand navbar-white px-0" do
          el.ul class: "align-items-center h-100 justify-content-around navbar-nav" do
            el.li class: "col nav-item px-0 text-center", "v-on:click.prevent": "tabBar.showSidebar" do
              el.a class: "text-dark", href: "/menu" do
                el.i(class: "fal fa-bars") {}
                el.div class: "small mt-1" do
                  t "noun.menu"
                end
              end
            end

            el.li class: "col nav-item px-0 text-center" do
              el.a class: "text-dark", href: "/" do
                el.i(class: "fal fa-home") {}
                el.div class: "mt-1 small" do
                  t "noun.home"
                end
              end
            end

            if user
              el.li class: "col nav-item px-0 text-center" do
                el.a class: "text-dark", href: "/programs" do
                  el.i(class: "fal fa-calendar") {}
                  el.div class: "mt-1 small" do
                    t "noun.slots"
                  end
                end
              end

              el.li class: "col nav-item px-0 text-center" do
                el.a class: "text-dark", href: "/@#{user.username}/watching" do
                  el.i(class: "fal fa-play") {}
                  el.div class: "mt-1 small" do
                    t "noun.watching_shorten"
                  end
                end
              end

              el.li class: "col nav-item px-0 text-center" do
                el.a class: "text-dark", href: "/works/#{ENV.fetch('ANNICT_CURRENT_SEASON')}" do
                  el.i(class: "fal fa-tv") {}
                  el.div class: "mt-1 small" do
                    t "noun.airing"
                  end
                end
              end
            else
              el.li class: "col nav-item px-0 text-center" do
                el.a class: "text-dark", href: "/works/#{ENV.fetch('ANNICT_CURRENT_SEASON')}" do
                  el.i(class: "fal fa-tv") {}
                  el.div class: "mt-1 small" do
                    t "noun.airing"
                  end
                end
              end

              el.li class: "col nav-item px-0 text-center" do
                el.a class: "text-dark", href: "/sign_up" do
                  el.i(class: "fal fa-rocket") {}
                  el.div class: "mt-1 small" do
                    t "noun.sign_up_shorten"
                  end
                end
              end

              el.li class: "col nav-item px-0 text-center" do
                el.a class: "text-dark", href: "/about" do
                  el.i(class: "fal fa-lightbulb") {}
                  el.div class: "mt-1 small" do
                    t "noun.about"
                  end
                end
              end
            end
          end
        end
      end
    end.html_safe
  end

  private

  attr_reader :user
end
