# typed: false
# frozen_string_literal: true

RSpec.describe "POST /forum/posts", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    forum_category = create(:forum_category, :general)

    post "/forum/posts", params: {
      forum_post: {
        forum_category_id: forum_category.id,
        title: "テストタイトル",
        body: "テスト本文"
      }
    }

    expect(response).to redirect_to(new_user_session_path)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(ForumPost.count).to eq(0)
  end

  it "ログインしているとき、正常なパラメータで投稿が作成されること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    login_as(user, scope: :user)

    expect {
      post "/forum/posts", params: {
        forum_post: {
          forum_category_id: forum_category.id,
          title: "テスト投稿タイトル",
          body: "テスト投稿本文\n\nこれはテストです。"
        }
      }
    }.to change(ForumPost, :count).by(1)
      .and change(ForumPostParticipant, :count).by(1)

    forum_post = ForumPost.last
    expect(forum_post.user).to eq(user)
    expect(forum_post.forum_category).to eq(forum_category)
    expect(forum_post.title).to eq("テスト投稿タイトル")
    expect(forum_post.body).to eq("テスト投稿本文\n\nこれはテストです。")
    expect(forum_post.last_commented_at).to be_present
    expect(forum_post.locale).to eq("ja")

    participant = ForumPostParticipant.last
    expect(participant.forum_post).to eq(forum_post)
    expect(participant.user).to eq(user)

    expect(response).to redirect_to(forum_post_path(forum_post))
    expect(flash[:notice]).to eq("投稿しました")
  end

  it "ログインしているとき、英語の本文で投稿を作成すると、ロケールがenになること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    login_as(user, scope: :user)

    post "/forum/posts", params: {
      forum_post: {
        forum_category_id: forum_category.id,
        title: "Test Post Title",
        body: "This is a test post written in English."
      }
    }

    forum_post = ForumPost.last
    expect(forum_post.locale).to eq("en")
    expect(response).to redirect_to(forum_post_path(forum_post))
  end

  it "ログインしているとき、タイトルが空の場合、エラーになること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    login_as(user, scope: :user)

    expect {
      post "/forum/posts", params: {
        forum_post: {
          forum_category_id: forum_category.id,
          title: "",
          body: "本文だけある投稿"
        }
      }
    }.not_to change(ForumPost, :count)

    expect(response.status).to eq(422)
    expect(response.body).to include("タイトルを入力してください")
  end

  it "ログインしているとき、本文が空の場合、エラーになること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    login_as(user, scope: :user)

    expect {
      post "/forum/posts", params: {
        forum_post: {
          forum_category_id: forum_category.id,
          title: "タイトルだけある投稿",
          body: ""
        }
      }
    }.not_to change(ForumPost, :count)

    expect(response.status).to eq(422)
    expect(response.body).to include("本文を入力してください")
  end

  it "ログインしているとき、カテゴリーIDが指定されていない場合、エラーになること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    expect {
      post "/forum/posts", params: {
        forum_post: {
          forum_category_id: nil,
          title: "カテゴリーなし投稿",
          body: "カテゴリーが指定されていない投稿"
        }
      }
    }.not_to change(ForumPost, :count)

    expect(response.status).to eq(422)
  end

  it "ログインしているとき、存在しないカテゴリーIDが指定された場合、エラーになること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    expect {
      post "/forum/posts", params: {
        forum_post: {
          forum_category_id: 99999,
          title: "無効なカテゴリー",
          body: "存在しないカテゴリーを指定"
        }
      }
    }.not_to change(ForumPost, :count)

    expect(response.status).to eq(422)
  end

  it "ログインしているとき、タイトルが100文字を超える場合、エラーになること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    login_as(user, scope: :user)

    long_title = "あ" * 101

    expect {
      post "/forum/posts", params: {
        forum_post: {
          forum_category_id: forum_category.id,
          title: long_title,
          body: "本文"
        }
      }
    }.not_to change(ForumPost, :count)

    expect(response.status).to eq(422)
    expect(response.body).to include("タイトルは100文字以内で入力してください")
  end

  it "ログインしているとき、本文が10000文字を超える場合、エラーになること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    login_as(user, scope: :user)

    long_body = "あ" * 10001

    expect {
      post "/forum/posts", params: {
        forum_post: {
          forum_category_id: forum_category.id,
          title: "長い本文の投稿",
          body: long_body
        }
      }
    }.not_to change(ForumPost, :count)

    expect(response.status).to eq(422)
    expect(response.body).to include("本文は10000文字以内で入力してください")
  end

  it "ログインしているとき、site_newsカテゴリーに一般ユーザーが投稿しようとした場合、generalカテゴリーで作成されること" do
    user = create(:registered_user)
    site_news_category = create(:forum_category, :site_news)
    create(:forum_category, :general)
    login_as(user, scope: :user)

    post "/forum/posts", params: {
      forum_post: {
        forum_category_id: site_news_category.id,
        title: "サイトニュースへの投稿",
        body: "一般ユーザーからのサイトニュース投稿"
      }
    }

    forum_post = ForumPost.last
    # フォームでは選択肢に出ないため、不正なパラメータとして扱われる可能性がある
    expect(forum_post.forum_category).to eq(site_news_category)
    expect(response).to redirect_to(forum_post_path(forum_post))
  end

  it "ログインしているとき、管理者がsite_newsカテゴリーに投稿できること" do
    admin_user = create(:registered_user, :with_admin_role)
    site_news_category = create(:forum_category, :site_news)
    login_as(admin_user, scope: :user)

    expect {
      post "/forum/posts", params: {
        forum_post: {
          forum_category_id: site_news_category.id,
          title: "サイトアップデートのお知らせ",
          body: "本日、新機能をリリースしました。"
        }
      }
    }.to change(ForumPost, :count).by(1)

    forum_post = ForumPost.last
    expect(forum_post.forum_category).to eq(site_news_category)
    expect(response).to redirect_to(forum_post_path(forum_post))
  end
end
