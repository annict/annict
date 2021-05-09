# frozen_string_literal: true

module Contents
  class AnimeRecordContentComponent < ApplicationComponent
    def initialize(view_context, record:, show_card: true)
      super view_context
      @record = record
      @anime_record = @record.anime_record
      @show_card = show_card
    end

    def render
      build_html do |h|
        h.tag :div do
          if @anime_record.rating_overall || @record_entity.body.present?
            h.html(SpoilerGuardComponent.new(view_context, record: @record).render { |h|
              h.tag :div, class: "mb-3 row" do
                h.tag :div, class: "col-12 col-xl-4 order-1 order-xl-2" do
                  if @anime_record.rating_overall
                    h.tag :div, class: "mb-3 mb-xl-0 p-3 rounded u-bg-black-000" do
                      h.tag :div, class: "small fw-bold text-center mb-2" do
                        h.text t("noun.rating")
                      end

                      if @anime_record.rating_animation
                        h.tag :div, class: "row" do
                          h.tag :div, class: "col" do
                            h.text t("noun.animation")
                          end

                          h.tag :div, class: "col pl-0 text-end" do
                            h.html RatingLabelComponent.new(view_context, rating: @anime_record.rating_animation).render
                          end
                        end
                      end

                      if @anime_record.rating_music
                        h.tag :div, class: "row" do
                          h.tag :div, class: "col" do
                            h.text t("noun.music")
                          end

                          h.tag :div, class: "col pl-0 text-end" do
                            h.html RatingLabelComponent.new(view_context, rating: @anime_record.rating_music).render
                          end
                        end
                      end

                      if @anime_record.rating_story
                        h.tag :div, class: "row" do
                          h.tag :div, class: "col" do
                            h.text t("noun.story")
                          end

                          h.tag :div, class: "col pl-0 text-end" do
                            h.html RatingLabelComponent.new(view_context, rating: @anime_record.rating_story).render
                          end
                        end
                      end

                      if @anime_record.rating_character
                        h.tag :div, class: "row" do
                          h.tag :div, class: "col" do
                            h.text t("noun.character")
                          end

                          h.tag :div, class: "col pl-0 text-end" do
                            h.html RatingLabelComponent.new(view_context, rating: @anime_record.rating_character).render
                          end
                        end
                      end

                      if @anime_record.rating_overall
                        h.tag :div, class: "row" do
                          h.tag :div, class: "col" do
                            h.text t("noun.overall")
                          end

                          h.tag :div, class: "col pl-0 text-end" do
                            h.html RatingLabelComponent.new(view_context, rating: @anime_record.rating_overall).render
                          end
                        end
                      end
                    end
                  end
                end

                h.tag :div, class: "col-12 col-xl-8 order-2 order-xl-1" do
                  h.html(BodyComponent.new(view_context, height: 300, format: :html).render { |h|
                    h.html render_markdown(@anime_record.comment)
                  })
                end
              end
            })
          end

          if @show_card
            h.html Cards::AnimeRecordCardComponent.new(view_context, anime_record: @anime_record).render
          end

          h.tag :div, class: "mt-1" do
            h.html Footers::RecordFooterComponent.new(view_context, record: @record).render
          end
        end
      end
    end
  end
end
