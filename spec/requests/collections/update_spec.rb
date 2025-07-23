# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /collections/:collection_id", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされること" do
    user = create(:registered_user)
    collection = create(:collection, user: user)

    patch "/collections/#{collection.id}", params: {
      forms_collection_form: {
        name: "更新されたコレクション名",
        description: "更新された説明文"
      }
    }

    expect(response).to redirect_to(new_user_session_path)
  end

  it "ログインしているとき、有効なパラメータでコレクションが更新されること" do
    user = create(:registered_user)
    collection = create(:collection, user: user, name: "元のコレクション名", description: "元の説明文")
    login_as(user, scope: :user)

    patch "/collections/#{collection.id}", params: {
      forms_collection_form: {
        name: "更新されたコレクション名",
        description: "更新された説明文"
      }
    }

    collection.reload
    expect(collection.name).to eq("更新されたコレクション名")
    expect(collection.description).to eq("更新された説明文")
    expect(response.status).to eq(302)
    expect(response).to redirect_to(user_collection_path(user.username, collection.id))
    expect(flash[:notice]).to eq(I18n.t("messages._common.updated"))
  end

  it "ログインしているとき、コレクション名のみ更新されること" do
    user = create(:registered_user)
    collection = create(:collection, user: user, name: "元のコレクション名", description: "元の説明文")
    login_as(user, scope: :user)

    patch "/collections/#{collection.id}", params: {
      forms_collection_form: {
        name: "更新されたコレクション名",
        description: "元の説明文"
      }
    }

    collection.reload
    expect(collection.name).to eq("更新されたコレクション名")
    expect(collection.description).to eq("元の説明文")
    expect(response.status).to eq(302)
    expect(response).to redirect_to(user_collection_path(user.username, collection.id))
    expect(flash[:notice]).to eq(I18n.t("messages._common.updated"))
  end

  it "ログインしているとき、説明文のみ更新されること" do
    user = create(:registered_user)
    collection = create(:collection, user: user, name: "元のコレクション名", description: "元の説明文")
    login_as(user, scope: :user)

    patch "/collections/#{collection.id}", params: {
      forms_collection_form: {
        name: "元のコレクション名",
        description: "更新された説明文"
      }
    }

    collection.reload
    expect(collection.name).to eq("元のコレクション名")
    expect(collection.description).to eq("更新された説明文")
    expect(response.status).to eq(302)
    expect(response).to redirect_to(user_collection_path(user.username, collection.id))
    expect(flash[:notice]).to eq(I18n.t("messages._common.updated"))
  end

  it "ログインしているとき、空の説明文でコレクションが更新されること" do
    user = create(:registered_user)
    collection = create(:collection, user: user, name: "元のコレクション名", description: "元の説明文")
    login_as(user, scope: :user)

    patch "/collections/#{collection.id}", params: {
      forms_collection_form: {
        name: "更新されたコレクション名",
        description: ""
      }
    }

    collection.reload
    expect(collection.name).to eq("更新されたコレクション名")
    expect(collection.description).to eq("")
    expect(response.status).to eq(302)
    expect(response).to redirect_to(user_collection_path(user.username, collection.id))
    expect(flash[:notice]).to eq(I18n.t("messages._common.updated"))
  end

  it "ログインしているとき、コレクション名が空の場合にバリデーションエラーになること" do
    user = create(:registered_user)
    collection = create(:collection, user: user, name: "元のコレクション名", description: "元の説明文")
    login_as(user, scope: :user)

    patch "/collections/#{collection.id}", params: {
      forms_collection_form: {
        name: "",
        description: "更新された説明文"
      }
    }

    collection.reload
    expect(collection.name).to eq("元のコレクション名")
    expect(collection.description).to eq("元の説明文")
    expect(response.status).to eq(422)
  end

  it "ログインしているとき、コレクション名が50文字を超える場合にバリデーションエラーになること" do
    user = create(:registered_user)
    collection = create(:collection, user: user, name: "元のコレクション名", description: "元の説明文")
    login_as(user, scope: :user)

    patch "/collections/#{collection.id}", params: {
      forms_collection_form: {
        name: "a" * 51,
        description: "更新された説明文"
      }
    }

    collection.reload
    expect(collection.name).to eq("元のコレクション名")
    expect(collection.description).to eq("元の説明文")
    expect(response.status).to eq(422)
  end

  it "ログインしているとき、説明文が最大文字数を超える場合にバリデーションエラーになること" do
    user = create(:registered_user)
    collection = create(:collection, user: user, name: "元のコレクション名", description: "元の説明文")
    login_as(user, scope: :user)

    patch "/collections/#{collection.id}", params: {
      forms_collection_form: {
        name: "更新されたコレクション名",
        description: "a" * 1_048_597
      }
    }

    collection.reload
    expect(collection.name).to eq("元のコレクション名")
    expect(collection.description).to eq("元の説明文")
    expect(response.status).to eq(422)
  end

  it "ログインしているとき、他人のコレクションを更新しようとすると、RecordNotFoundエラーが発生すること" do
    user = create(:registered_user)
    other_user = create(:registered_user)
    collection = create(:collection, user: other_user)
    login_as(user, scope: :user)

    expect {
      patch "/collections/#{collection.id}", params: {
        forms_collection_form: {
          name: "更新されたコレクション名",
          description: "更新された説明文"
        }
      }
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "ログインしているとき、存在しないコレクションを更新しようとすると、RecordNotFoundエラーが発生すること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    expect {
      patch "/collections/nonexistent", params: {
        forms_collection_form: {
          name: "更新されたコレクション名",
          description: "更新された説明文"
        }
      }
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "ログインしているとき、削除済みのコレクションを更新しようとすると、RecordNotFoundエラーが発生すること" do
    user = create(:registered_user)
    collection = create(:collection, user: user, deleted_at: Time.current)
    login_as(user, scope: :user)

    expect {
      patch "/collections/#{collection.id}", params: {
        forms_collection_form: {
          name: "更新されたコレクション名",
          description: "更新された説明文"
        }
      }
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
