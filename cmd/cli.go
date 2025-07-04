package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/VladMinzatu/reference-manager/adapters"
	"github.com/VladMinzatu/reference-manager/domain/model"
	"github.com/VladMinzatu/reference-manager/domain/service"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "refman",
	Short: "Reference Manager CLI",
	Long:  "Command line interface for managing references organized by categories",
}

func main() {
	db, err := sql.Open("sqlite3", "db/references.db")
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}
	defer db.Close()

	categoryRepo := adapters.NewSQLiteCategoryRepository(db)
	categoryService := service.NewCategoryService(categoryRepo)
	categoryListRepository := adapters.NewSQLiteCategoryListRepository(db)
	referenceRepo := adapters.NewSQLiteReferencesRepository(db)

	// Category commands
	var categoryCmd = &cobra.Command{
		Use:   "category",
		Short: "Manage categories",
	}

	var addCategoryCmd = &cobra.Command{
		Use:   "add [name]",
		Short: "Add a new category",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title, err := model.NewTitle(args[0])
			if err != nil {
				return fmt.Errorf("invalid category name: %v", err)
			}
			cat, err := categoryListRepository.AddNewCategory(title)
			if err != nil {
				return err
			}
			fmt.Printf("Added category: %s (id: %d)\n", cat.Name, cat.Id)
			return nil
		},
	}

	var listCategoriesCmd = &cobra.Command{
		Use:   "list",
		Short: "List all categories",
		RunE: func(cmd *cobra.Command, args []string) error {
			categories, err := categoryListRepository.GetAllCategoryRefs()
			if err != nil {
				return err
			}
			for _, cat := range categories {
				fmt.Printf("%d: %s\n", cat.Id, cat.Name)
			}
			return nil
		},
	}

	var updateCategoryCmd = &cobra.Command{
		Use:   "update [id] [new_name]",
		Short: "Update the name of a category",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid category id format (must be integer): %v", err)
			}
			catId, err := model.NewId(id)
			if err != nil {
				return fmt.Errorf("invalid category id: %v", err)
			}
			newName, err := model.NewTitle(args[1])
			if err != nil {
				return fmt.Errorf("invalid category name: %v", err)
			}
			if _, err := categoryService.UpdateTitle(catId, newName); err != nil {
				return err
			}
			fmt.Printf("Updated category %d to name: %s\n", id, newName)
			return nil
		},
	}

	var deleteCategoryCmd = &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a category",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid category id format (must be integer): %v", err)
			}
			modelId, err := model.NewId(id)
			if err != nil {
				return fmt.Errorf("invalid category id: %v", err)
			}
			if err := categoryListRepository.DeleteCategory(modelId); err != nil {
				return err
			}
			fmt.Printf("Deleted category with id: %d\n", id)
			return nil
		},
	}

	var reorderCategoriesCmd = &cobra.Command{
		Use:   "reorder [id1] [id2] ...",
		Short: "Reorder categories by specifying their ids in the desired order",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			positions := make(map[model.Id]int)
			for pos, arg := range args {
				id, err := strconv.ParseInt(arg, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid category id: %s", arg)
				}
				modelId, err := model.NewId(id)
				if err != nil {
					return fmt.Errorf("invalid category id: %s", arg)
				}
				positions[modelId] = pos
			}
			if err := categoryListRepository.ReorderCategories(positions); err != nil {
				return err
			}
			fmt.Println("Categories reordered successfully.")
			return nil
		},
	}

	// Reference commands
	var referenceCmd = &cobra.Command{
		Use:   "reference",
		Short: "Manage references",
	}

	var listReferencesCmd = &cobra.Command{
		Use:   "list [categoryId] [starredOnly]",
		Short: "List references in a category, optionally filtering by starred references",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid category id: %v", err)
			}
			catId, err := model.NewId(id)
			if err != nil {
				return fmt.Errorf("invalid category id: %v", err)
			}
			category, err := categoryService.GetCategoryById(catId)
			if err != nil {
				return err
			}
			for _, ref := range category.References {
				ref.Render(&CLIReferenceRenderer{})
				fmt.Println()
			}
			return nil
		},
	}

	var addBookCmd = &cobra.Command{
		Use:   "add-book [categoryId] [title] [isbn] [description]",
		Short: "Add a book reference",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			categoryId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid category id: %v", err)
			}
			catId, err := model.NewId(categoryId)
			if err != nil {
				return fmt.Errorf("invalid category id: %v", err)
			}
			title, err := model.NewTitle(args[1])
			if err != nil {
				return fmt.Errorf("invalid title: %v", err)
			}
			isbn, err := model.NewISBN(args[2])
			if err != nil {
				return fmt.Errorf("invalid ISBN: %v", err)
			}
			description := args[3]
			// Book id will be assigned by the system, so we use a placeholder zero value for id here
			book := model.NewBookReference(0, title, isbn, description, false)
			category, err := categoryService.AddReference(catId, book)
			if err != nil {
				return err
			}
			addedBook := category.References[len(category.References)-1]
			fmt.Printf("Added book: %s (id: %d)\n", addedBook.Title(), addedBook.GetId())
			return nil
		},
	}

	var updateBookCmd = &cobra.Command{
		Use:   "update-book [id] [title] [isbn] [description] [starred]",
		Short: "Update a book reference",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			idInt, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid book id: %v", err)
			}
			bookId, err := model.NewId(idInt)
			if err != nil {
				return fmt.Errorf("invalid book id: %v", err)
			}
			title, err := model.NewTitle(args[1])
			if err != nil {
				return fmt.Errorf("invalid title: %v", err)
			}
			isbn, err := model.NewISBN(args[2])
			if err != nil {
				return fmt.Errorf("invalid ISBN: %v", err)
			}
			description := args[3]
			starred, err := strconv.ParseBool(args[4])
			if err != nil {
				return fmt.Errorf("invalid starred value (must be true or false): %v", err)
			}
			// Construct the updated book reference
			updatedBook := model.NewBookReference(bookId, title, isbn, description, starred)
			if err := referenceRepo.UpdateReference(bookId, updatedBook); err != nil {
				return err
			}
			return nil
		},
	}

	var addLinkCmd = &cobra.Command{
		Use:   "add-link [categoryId] [title] [url] [description]",
		Short: "Add a link reference",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			catIdInt, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid category id: %v", err)
			}
			catId, err := model.NewId(catIdInt)
			if err != nil {
				return fmt.Errorf("invalid category id: %v", err)
			}
			title, err := model.NewTitle(args[1])
			if err != nil {
				return fmt.Errorf("invalid title: %v", err)
			}
			url, err := model.NewURL(args[2])
			if err != nil {
				return fmt.Errorf("invalid URL: %v", err)
			}
			description := args[3]
			// Link id will be assigned by the system, so we use a placeholder zero value for id here
			link := model.NewLinkReference(0, title, url, description, false)
			category, err := categoryService.AddReference(catId, link)
			if err != nil {
				return err
			}
			addedLink := category.References[len(category.References)-1]
			fmt.Printf("Added link: %s (id: %d)\n", addedLink.Title(), addedLink.GetId())
			return nil
		},
	}

	var updateLinkCmd = &cobra.Command{
		Use:   "update-link [id] [title] [url] [description] [starred]",
		Short: "Update a link reference",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			idInt, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid link id: %v", err)
			}
			linkId, err := model.NewId(idInt)
			if err != nil {
				return fmt.Errorf("invalid link id: %v", err)
			}
			title, err := model.NewTitle(args[1])
			if err != nil {
				return fmt.Errorf("invalid title: %v", err)
			}
			url, err := model.NewURL(args[2])
			if err != nil {
				return fmt.Errorf("invalid URL: %v", err)
			}
			description := args[3]
			starred, err := strconv.ParseBool(args[4])
			if err != nil {
				return fmt.Errorf("invalid starred value (must be true or false): %v", err)
			}
			updatedLink := model.NewLinkReference(linkId, title, url, description, starred)
			if err := referenceRepo.UpdateReference(linkId, updatedLink); err != nil {
				return err
			}
			fmt.Printf("Updated link (id: %d)\n", linkId)
			return nil
		},
	}

	var addNoteCmd = &cobra.Command{
		Use:   "add-note [categoryId] [title] [text]",
		Short: "Add a note reference",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			catIdInt, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid category id: %v", err)
			}
			catId, err := model.NewId(catIdInt)
			if err != nil {
				return fmt.Errorf("invalid category id: %v", err)
			}
			title, err := model.NewTitle(args[1])
			if err != nil {
				return fmt.Errorf("invalid title: %v", err)
			}
			text := args[2]
			// Note id will be assigned by the system, so we use a placeholder zero value for id here
			note := model.NewNoteReference(0, title, text, false)
			category, err := categoryService.AddReference(catId, note)
			if err != nil {
				return err
			}
			addedNote := category.References[len(category.References)-1]
			fmt.Printf("Added note: %s (id: %d)\n", addedNote.Title(), addedNote.GetId())
			return nil
		},
	}

	var updateNoteCmd = &cobra.Command{
		Use:   "update-note [id] [title] [text] [starred]",
		Short: "Update a note reference",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			idInt, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid note id: %v", err)
			}
			noteId, err := model.NewId(idInt)
			if err != nil {
				return fmt.Errorf("invalid note id: %v", err)
			}
			title, err := model.NewTitle(args[1])
			if err != nil {
				return fmt.Errorf("invalid title: %v", err)
			}
			text := args[2]
			starred, err := strconv.ParseBool(args[3])
			if err != nil {
				return fmt.Errorf("invalid starred value (must be true or false): %v", err)
			}
			updatedNote := model.NewNoteReference(noteId, title, text, starred)
			if err := referenceRepo.UpdateReference(noteId, updatedNote); err != nil {
				return err
			}
			fmt.Printf("Updated note (id: %d)\n", noteId)
			return nil
		},
	}

	var deleteReferenceCmd = &cobra.Command{
		Use:   "delete [category_id] [reference_id]",
		Short: "Delete a reference from a category",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			categoryIdInt, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid category id: %v", err)
			}
			catId, err := model.NewId(categoryIdInt)
			if err != nil {
				return fmt.Errorf("invalid category id: %v", err)
			}
			refIdInt, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid reference id: %v", err)
			}
			refId, err := model.NewId(refIdInt)
			if err != nil {
				return fmt.Errorf("invalid reference id: %v", err)
			}
			if _, err := categoryService.RemoveReference(catId, refId); err != nil {
				return err
			}
			fmt.Printf("Deleted reference with id: %d from category: %d\n", refIdInt, categoryIdInt)
			return nil
		},
	}

	var reorderReferencesCmd = &cobra.Command{
		Use:   "reorder [categoryId] [id1] [id2] ...",
		Short: "Reorder references in a category by specifying the ids in the desired order.",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			categoryIdInt, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid category id: %v", err)
			}
			categoryId, err := model.NewId(categoryIdInt)
			if err != nil {
				return fmt.Errorf("invalid category id: %v", err)
			}
			positions := make(map[model.Id]int)
			for pos, idStr := range args[1:] {
				idInt, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid reference id at position %d: %v", pos, err)
				}
				id, err := model.NewId(idInt)
				if err != nil {
					return fmt.Errorf("invalid reference id at position %d: %v", pos, err)
				}
				positions[id] = pos
			}
			_, err = categoryService.ReorderReferences(categoryId, positions)
			if err != nil {
				return err
			}
			fmt.Printf("Reordered references in category %d\n", categoryId)
			return nil
		},
	}

	categoryCmd.AddCommand(addCategoryCmd, listCategoriesCmd, updateCategoryCmd, deleteCategoryCmd, reorderCategoriesCmd)
	referenceCmd.AddCommand(listReferencesCmd, addBookCmd, updateBookCmd, addLinkCmd, updateLinkCmd, addNoteCmd, updateNoteCmd, deleteReferenceCmd, reorderReferencesCmd)
	rootCmd.AddCommand(categoryCmd, referenceCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type CLIReferenceRenderer struct{}

func (r *CLIReferenceRenderer) RenderBook(ref model.BookReference) {
	fmt.Printf("%d: %s [Book] %s\n", ref.GetId(), r.StarChar(ref.Starred()), ref.Title())
	fmt.Printf("\t\t\tISBN: %s\n", ref.ISBN)
	fmt.Printf("\t\t\tDescription: %s\n", ref.Description)
}

func (r *CLIReferenceRenderer) RenderLink(ref model.LinkReference) {
	fmt.Printf("%d: %s [Link] %s\n", ref.GetId(), r.StarChar(ref.Starred()), ref.Title())
	fmt.Printf("\t\t\tURL: %s\n", ref.URL)
	fmt.Printf("\t\t\tDescription: %s\n", ref.Description)
}

func (r *CLIReferenceRenderer) RenderNote(ref model.NoteReference) {
	fmt.Printf("%d: %s [Note] %s\n", ref.GetId(), r.StarChar(ref.Starred()), ref.Title())
	fmt.Printf("\t\t\tText: %s\n", ref.Text)
}

func (r *CLIReferenceRenderer) StarChar(starred bool) string {
	if starred {
		return "★"
	}
	return "☆"
}
