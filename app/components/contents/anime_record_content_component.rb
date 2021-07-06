# frozen_string_literal: true

module Contents
  class AnimeRecordContentComponent < ApplicationV6Component
    def initialize(view_context, record:, show_box: true)
      super view_context
      @record = record
      @anime_record = @record.anime_record
      @show_box = show_box
      @anime = @record.anime
    end

    def render
      build_html do |h|
        h.tag :div do
          if @anime_record.rating_overall_state || @anime_record.comment.present?
            h.html(SpoilerGuardV6Component.new(view_context, record: @record).render { |h|
              h.tag :div, class: "mb-3 row" do
                h.tag :div, class: "col-12 col-xl-4 order-1 order-xl-2" do
                  if @anime_record.rating_overall_state
                    h.tag :div, class: "mb-3 mb-xl-0 p-3 rounded u-bg-black-000" do
                      h.tag :div, class: "small fw-bold text-center mb-2" do
                        h.text t("noun.rating")
                      end

                      if @anime_record.rating_animation_state
                        h.tag :div, class: "row" do
                          h.tag :div, class: "col" do
                            h.text t("noun.animation")
                          end

                          h.tag :div, class: "col ps-0 text-end" do
                            h.html Labels::RatingLabelComponent.new(view_context, rating: @anime_record.rating_animation_state).render
                          end
                        end
                      end

                      if @anime_record.rating_music_state
                        h.tag :div, class: "row" do
                          h.tag :div, class: "col" do
                            h.text t("noun.music")
                          end

                          h.tag :div, class: "col ps-0 text-end" do
                            h.html Labels::RatingLabelComponent.new(view_context, rating: @anime_record.rating_music_state).render
                          end
                        end
                      end

                      if @anime_record.rating_story_state
                        h.tag :div, class: "row" do
                          h.tag :div, class: "col" do
                            h.text t("noun.story")
                          end

                          h.tag :div, class: "col ps-0 text-end" do
                            h.html Labels::RatingLabelComponent.new(view_context, rating: @anime_record.rating_story_state).render
                          end
                        end
                      end

                      if @anime_record.rating_character_state
                        h.tag :div, class: "row" do
                          h.tag :div, class: "col" do
                            h.text t("noun.character")
                          end

                          h.tag :div, class: "col ps-0 text-end" do
                            h.html Labels::RatingLabelComponent.new(view_context, rating: @anime_record.rating_character_state).render
                          end
                        end
                      end

                      if @anime_record.rating_overall_state
                        h.tag :div, class: "row" do
                          h.tag :div, class: "col" do
                            h.text t("noun.overall")
                          end

                          h.tag :div, class: "col ps-0 text-end" do
                            h.html Labels::RatingLabelComponent.new(view_context, rating: @anime_record.rating_overall_state).render
                          end
                        end
                      end
                    end
                  end
                end

                h.tag :div, class: "col-12 col-xl-8 order-2 order-xl-1" do
                  h.html BodyV6Component.new(view_context, content: @anime_record.comment, format: :markdown, height: 300).render
                end
              end
            })
          end

          if @show_box
            h.tag :hr
            h.html Boxes::AnimeBoxComponent.new(view_context, anime: @anime).render
            h.tag :hr
          end

          h.tag :div, class: "mt-1" do
            h.html Footers::RecordFooterComponent.new(view_context, record: @record).render
          end
        end
      end
    end
  end
end
