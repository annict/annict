# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /forum/posts/:post_id", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user:, forum_category:,
      title: "元のタイトル", body: "元の本文")

    patch "/forum/posts/#{forum_post.id}", params: {
      forum_post: {
        forum_category_id: forum_category.id,
        title: "更新後のタイトル",
        body: "更新後の本文"
      }
    }

    expect(response).to redirect_to(new_user_session_path)
    expect(flash[:alert]).to eq("ログインしてください")
    forum_post.reload
    expect(forum_post.title).to eq("元のタイトル")
    expect(forum_post.body).to eq("元の本文")
  end

  it "ログインしているとき、自分の投稿を正常に更新できること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user:, forum_category:,
      title: "元のタイトル", body: "元の本文", locale: "ja")
    login_as(user, scope: :user)

    patch "/forum/posts/#{forum_post.id}", params: {
      forum_post: {
        forum_category_id: forum_category.id,
        title: "更新後のタイトル",
        body: "更新後の本文\n\nこれは更新テストです。"
      }
    }

    forum_post.reload
    expect(forum_post.title).to eq("更新後のタイトル")
    expect(forum_post.body).to eq("更新後の本文\n\nこれは更新テストです。")
    expect(forum_post.locale).to eq("ja")
    expect(response).to redirect_to(forum_post_path(forum_post))
    expect(flash[:notice]).to eq("更新しました")
  end

  it "ログインしているとき、英語の本文に更新すると、ロケールがenになること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user:, forum_category:,
      title: "日本語タイトル", body: "日本語の本文です", locale: "ja")
    login_as(user, scope: :user)

    patch "/forum/posts/#{forum_post.id}", params: {
      forum_post: {
        forum_category_id: forum_category.id,
        title: "Updated Title",
        body: "This is an updated post written in English."
      }
    }

    forum_post.reload
    expect(forum_post.locale).to eq("en")
    expect(response).to redirect_to(forum_post_path(forum_post))
  end

  it "ログインしているとき、他人の投稿を更新しようとすると、403エラーになること" do
    user = create(:registered_user)
    other_user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user: other_user, forum_category:,
      title: "他人の投稿", body: "他人の投稿本文")
    login_as(user, scope: :user)

    expect {
      patch "/forum/posts/#{forum_post.id}", params: {
        forum_post: {
          forum_category_id: forum_category.id,
          title: "更新しようとしたタイトル",
          body: "更新しようとした本文"
        }
      }
    }.to raise_error(Pundit::NotAuthorizedError)

    forum_post.reload
    expect(forum_post.title).to eq("他人の投稿")
    expect(forum_post.body).to eq("他人の投稿本文")
  end

  it "ログインしているとき、タイトルが空の場合、エラーになること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user:, forum_category:,
      title: "元のタイトル", body: "元の本文")
    login_as(user, scope: :user)

    patch "/forum/posts/#{forum_post.id}", params: {
      forum_post: {
        forum_category_id: forum_category.id,
        title: "",
        body: "本文はある"
      }
    }

    expect(response.status).to eq(422)
    expect(response.body).to include("タイトルを入力してください")
    forum_post.reload
    expect(forum_post.title).to eq("元のタイトル")
  end

  it "ログインしているとき、本文が空の場合、エラーになること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user:, forum_category:,
      title: "元のタイトル", body: "元の本文")
    login_as(user, scope: :user)

    patch "/forum/posts/#{forum_post.id}", params: {
      forum_post: {
        forum_category_id: forum_category.id,
        title: "タイトルはある",
        body: ""
      }
    }

    expect(response.status).to eq(422)
    expect(response.body).to include("本文を入力してください")
    forum_post.reload
    expect(forum_post.body).to eq("元の本文")
  end

  it "ログインしているとき、タイトルが100文字を超える場合、エラーになること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user:, forum_category:,
      title: "元のタイトル", body: "元の本文")
    login_as(user, scope: :user)

    long_title = "あ" * 101

    patch "/forum/posts/#{forum_post.id}", params: {
      forum_post: {
        forum_category_id: forum_category.id,
        title: long_title,
        body: "本文"
      }
    }

    expect(response.status).to eq(422)
    expect(response.body).to include("タイトルは100文字以内で入力してください")
    forum_post.reload
    expect(forum_post.title).to eq("元のタイトル")
  end

  it "ログインしているとき、本文が10000文字を超える場合、エラーになること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user:, forum_category:,
      title: "元のタイトル", body: "元の本文")
    login_as(user, scope: :user)

    long_body = "あ" * 10001

    patch "/forum/posts/#{forum_post.id}", params: {
      forum_post: {
        forum_category_id: forum_category.id,
        title: "タイトル",
        body: long_body
      }
    }

    expect(response.status).to eq(422)
    expect(response.body).to include("本文は10000文字以内で入力してください")
    forum_post.reload
    expect(forum_post.body).to eq("元の本文")
  end

  it "ログインしているとき、カテゴリーを変更できること" do
    user = create(:registered_user)
    forum_category1 = create(:forum_category, :general)
    forum_category2 = create(:forum_category, :feedback)
    forum_post = create(:forum_post, user:, forum_category: forum_category1,
      title: "タイトル", body: "本文")
    login_as(user, scope: :user)

    patch "/forum/posts/#{forum_post.id}", params: {
      forum_post: {
        forum_category_id: forum_category2.id,
        title: "タイトル",
        body: "本文"
      }
    }

    forum_post.reload
    expect(forum_post.forum_category).to eq(forum_category2)
    expect(response).to redirect_to(forum_post_path(forum_post))
  end

  it "ログインしているとき、存在しないカテゴリーIDが指定された場合、エラーになること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    forum_post = create(:forum_post, user:, forum_category:,
      title: "元のタイトル", body: "元の本文")
    login_as(user, scope: :user)

    patch "/forum/posts/#{forum_post.id}", params: {
      forum_post: {
        forum_category_id: 99999,
        title: "タイトル",
        body: "本文"
      }
    }

    expect(response.status).to eq(422)
    forum_post.reload
    expect(forum_post.forum_category).to eq(forum_category)
  end

  it "ログインしているとき、存在しない投稿を更新しようとすると、404エラーになること" do
    user = create(:registered_user)
    forum_category = create(:forum_category, :general)
    login_as(user, scope: :user)

    patch "/forum/posts/99999", params: {
    forum_post: {
    forum_category_id: forum_category.id,
    title: "タイトル",
    body: "本文"

    expect(response.status).to eq(404)
  end
end
